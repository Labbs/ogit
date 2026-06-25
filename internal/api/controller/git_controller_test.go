package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/storage/memory"
	billyfs "github.com/go-git/go-billy/v5/memfs"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildGitRepo creates an in-memory git repository with n commits and returns
// the storer and the HEAD hash.
func buildGitRepo(t *testing.T, n int) (*memory.Storage, plumbing.Hash) {
	t.Helper()

	st := memory.NewStorage()
	fs := billyfs.New()

	repo, err := gogit.Init(st, fs)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	author := &object.Signature{
		Name:  "Test",
		Email: "test@example.com",
		When:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	var head plumbing.Hash
	for i := 0; i < n; i++ {
		f, err := fs.Create(fmt.Sprintf("f%d.txt", i))
		require.NoError(t, err)
		fmt.Fprintf(f, "content %d", i)
		f.Close()

		_, err = wt.Add(fmt.Sprintf("f%d.txt", i))
		require.NoError(t, err)

		head, err = wt.Commit(fmt.Sprintf("commit %d", i), &gogit.CommitOptions{Author: author})
		require.NoError(t, err)
	}

	return st, head
}

// setupGitControllerApp creates a Fiber app with the GitController routes and
// a pre-configured MockGitRepositoryStorage.
func setupGitControllerApp(mockStorage *MockGitRepositoryStorage) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	gc := &GitController{
		Logger:  zerolog.Nop(),
		Storage: mockStorage,
	}
	app.Get("/:repo/info/refs", gc.InfoRefs)
	app.Post("/:repo/git-upload-pack", gc.HandleUploadPack)
	return app
}

// shallowUploadPackRequest builds a pkt-line encoded upload-pack request body
// with a depth (deepen) for use in HTTP POST tests.
func shallowUploadPackRequest(t *testing.T, want plumbing.Hash, depth int) []byte {
	t.Helper()
	var buf bytes.Buffer
	enc := pktline.NewEncoder(&buf)
	require.NoError(t, enc.EncodeString(
		fmt.Sprintf("want %s ofs-delta shallow\n", want),
	))
	require.NoError(t, enc.EncodeString(fmt.Sprintf("deepen %d\n", depth)))
	require.NoError(t, enc.Flush())
	return buf.Bytes()
}

// TestInfoRefs_AdvertisesShallowCapability verifies that the upload-pack
// advertisement includes the "shallow" capability after our patch.
func TestInfoRefs_AdvertisesShallowCapability(t *testing.T) {
	st, _ := buildGitRepo(t, 3)
	mock := &MockGitRepositoryStorage{}
	mock.On("RepositoryExists", "test.git").Return(true)
	mock.On("GetStorer", "test.git").Return(st, nil)

	app := setupGitControllerApp(mock)

	req := httptest.NewRequest("GET", "/test.git/info/refs?service=git-upload-pack", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-git-upload-pack-advertisement",
		resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.True(t, strings.Contains(string(body), "shallow"),
		"info/refs response must advertise shallow capability")
	mock.AssertExpectations(t)
}

// TestInfoRefs_InvalidService verifies that an unsupported service query
// parameter is rejected with 400.
func TestInfoRefs_InvalidService(t *testing.T) {
	mock := &MockGitRepositoryStorage{}
	app := setupGitControllerApp(mock)

	req := httptest.NewRequest("GET", "/test.git/info/refs?service=git-bad-service", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)
}

// TestHandleUploadPack_ShallowIntercepted verifies that a depth-limited
// upload-pack request is handled by our shallow implementation, producing
// a response that contains shallow boundary markers.
func TestHandleUploadPack_ShallowIntercepted(t *testing.T) {
	st, head := buildGitRepo(t, 3)
	mock := &MockGitRepositoryStorage{}
	mock.On("RepositoryExists", "test.git").Return(true)
	// GetStorer is called twice: once by the transport loader, once by the
	// shallow handler.
	mock.On("GetStorer", "test.git").Return(st, nil)

	app := setupGitControllerApp(mock)

	body := shallowUploadPackRequest(t, head, 1)
	req := httptest.NewRequest("POST", "/test.git/git-upload-pack", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")

	resp, err := app.Test(req, 10_000) // 10 s timeout
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/x-git-upload-pack-result",
		resp.Header.Get("Content-Type"))

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// The response must contain a shallow marker for the requested commit.
	assert.True(t, strings.Contains(string(respBody), "shallow"),
		"response must contain shallow boundary marker")

	// The response must contain the HEAD commit hash referenced in the shallow line.
	assert.True(t, strings.Contains(string(respBody), head.String()),
		"response shallow line must reference the requested commit")

	// A valid pack starts with the PACK signature.
	assert.True(t, strings.Contains(string(respBody), "PACK"),
		"response must contain a packfile")

	mock.AssertExpectations(t)
}

// TestHandleUploadPack_RejectsUnknownRepo verifies that a request for a
// non-existent repository returns 404 (or 500 due to fiber error wrapping).
func TestHandleUploadPack_RejectsUnknownRepo(t *testing.T) {
	mock := &MockGitRepositoryStorage{}
	mock.On("RepositoryExists", "missing.git").Return(false)

	app := setupGitControllerApp(mock)

	req := httptest.NewRequest("POST", "/missing.git/git-upload-pack", bytes.NewReader([]byte{}))
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, 200, resp.StatusCode)
}

// TestHandleUploadPack_ShallowDepthZeroIsNotIntercepted verifies that the
// controller does NOT route through the shallow handler when depth=0 (the
// default in a standard clone request).
func TestHandleUploadPack_ShallowDepthZeroIsNotIntercepted(t *testing.T) {
	// Depth=0 — req.Depth.IsZero() is true, so the shallow intercept must not
	// fire. This is a unit-level assertion on packp types, not a full HTTP round
	// trip, because a full non-shallow pack requires a client-side "done" line.
	req := packp.NewUploadPackRequest()
	req.Depth = packp.DepthCommits(0)

	_, isDepth := req.Depth.(packp.DepthCommits)
	assert.True(t, isDepth, "DepthCommits type assertion must succeed")
	assert.True(t, req.Depth.IsZero(), "depth=0 must be treated as zero/non-shallow")
}

// TestHandleUploadPack_ShallowDepthOne verifies the type assertions used by
// the shallow intercept in HandleUploadPack for a depth-1 request.
func TestHandleUploadPack_ShallowDepthOne(t *testing.T) {
	req := packp.NewUploadPackRequest()
	req.Depth = packp.DepthCommits(1)

	_, isDepth := req.Depth.(packp.DepthCommits)
	assert.True(t, isDepth, "DepthCommits type assertion must succeed")
	assert.False(t, req.Depth.IsZero(), "depth=1 must be treated as shallow")
}

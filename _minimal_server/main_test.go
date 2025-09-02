package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Test end-to-end: init server, push to a new repo, then ls-remote
func TestSmartHTTP_PushAndList(t *testing.T) {
	dir := t.TempDir()
	BaseDir = filepath.Join(dir, "repositories")
	if err := os.MkdirAll(BaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Start native Fiber server on ephemeral port
	app := NewGitHTTPApp()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	done := make(chan struct{})
	go func() {
		_ = app.Listener(ln)
		close(done)
	}()
	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = ln.Close()
		<-done
	})

	base := "http://" + ln.Addr().String()
	remoteURL := base + "/demo.git"

	// Create repo via API
	resp, err := http.Post(base+"/api/repos", "application/json", strings.NewReader(`{"name":"demo.git"}`))
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Fatalf("create repo status: %v", resp.Status)
	}

	// Create a temp repo and push
	work := filepath.Join(dir, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}

	r, err := gogit.PlainInit(work, false)
	if err != nil {
		t.Fatalf("init work: %v", err)
	}

	// Write a file and commit
	if err := os.WriteFile(filepath.Join(work, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	wt, err := r.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := wt.Commit("init", &gogit.CommitOptions{All: true, Author: defaultSig()}); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Add remote and push main
	if _, err := r.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{remoteURL}}); err != nil {
		t.Fatalf("create remote: %v", err)
	}
	if err := r.Push(&gogit.PushOptions{RemoteName: "origin"}); err != nil {
		t.Fatalf("push: %v", err)
	}

	// Now ls-remote via HTTP (GET info/refs upload-pack)
	resp, err = http.Get(remoteURL + "/info/refs?service=git-upload-pack")
	if err != nil {
		t.Fatalf("ls-remote http: %v", err)
	}
	t.Cleanup(func() { io.Copy(io.Discard, resp.Body); resp.Body.Close() })
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status: %v", resp.Status)
	}
}

func defaultSig() *object.Signature {
	// Minimal signature for tests
	return &object.Signature{Name: "tester", Email: "tester@example.com", When: time.Now()}
}

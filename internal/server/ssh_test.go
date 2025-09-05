package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/labbs/git-server-s3/pkg/storage/local"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHConfig_Configure(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "ssh-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup storage
	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	localStorage := local.NewLocalStorage(logger)
	localStorage.Configure()

	// Create SSH config
	sshConfig := &SSHConfig{
		Port:        ":0", // Use random port for testing
		Logger:      logger,
		Storage:     localStorage,
		HostKeyPath: filepath.Join(tempDir, "test_host_key"),
	}

	// Test configuration
	err = sshConfig.Configure()
	require.NoError(t, err)

	// Verify server was created
	assert.NotNil(t, sshConfig.Server)
	assert.Equal(t, ":0", sshConfig.Server.Addr)

	// Verify host key was generated
	assert.FileExists(t, sshConfig.HostKeyPath)
}

func TestSSHConfig_parseGitCommand(t *testing.T) {
	sshConfig := &SSHConfig{}

	tests := []struct {
		name        string
		command     string
		wantService string
		wantRepo    string
	}{
		{
			name:        "upload pack command",
			command:     "git-upload-pack '/repo.git'",
			wantService: "git-upload-pack",
			wantRepo:    "'/repo.git'",
		},
		{
			name:        "receive pack command",
			command:     "git-receive-pack '/repo.git'",
			wantService: "git-receive-pack",
			wantRepo:    "'/repo.git'",
		},
		{
			name:        "upload pack with spaces",
			command:     "git-upload-pack   '/my repo.git'  ",
			wantService: "git-upload-pack",
			wantRepo:    "'/my repo.git'",
		},
		{
			name:        "invalid command",
			command:     "invalid-command",
			wantService: "",
			wantRepo:    "",
		},
		{
			name:        "empty command",
			command:     "",
			wantService: "",
			wantRepo:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, repo := sshConfig.parseGitCommand(tt.command)
			assert.Equal(t, tt.wantService, service)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}

func TestSSHConfig_repoPathFromSSHArg(t *testing.T) {
	sshConfig := &SSHConfig{}

	tests := []struct {
		name     string
		arg      string
		expected string
	}{
		{
			name:     "simple repo path",
			arg:      "'repo.git'",
			expected: "repo.git",
		},
		{
			name:     "double quoted path",
			arg:      "\"repo.git\"",
			expected: "repo.git",
		},
		{
			name:     "path with leading slash",
			arg:      "'/repo.git'",
			expected: "repo.git",
		},
		{
			name:     "path with colon prefix",
			arg:      "':repo.git'",
			expected: "repo.git",
		},
		{
			name:     "path with host",
			arg:      "'host:/path/repo.git'",
			expected: "path/repo.git",
		},
		{
			name:     "repo without .git suffix",
			arg:      "'myrepo'",
			expected: "myrepo.git",
		},
		{
			name:     "empty path",
			arg:      "''",
			expected: "",
		},
		{
			name:     "path with spaces",
			arg:      "  '/repo.git'  ",
			expected: "repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sshConfig.repoPathFromSSHArg(tt.arg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSSHConfig_ensureHostKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hostkey-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	sshConfig := &SSHConfig{
		Logger:      logger,
		HostKeyPath: filepath.Join(tempDir, "host_key"),
	}

	// Test key generation
	signer1, err := sshConfig.ensureHostKey()
	require.NoError(t, err)
	assert.NotNil(t, signer1)

	// Verify key file was created
	assert.FileExists(t, sshConfig.HostKeyPath)

	// Test key loading (should load the same key)
	signer2, err := sshConfig.ensureHostKey()
	require.NoError(t, err)
	assert.NotNil(t, signer2)

	// Both signers should have the same key
	assert.Equal(t, signer1.PublicKey().Marshal(), signer2.PublicKey().Marshal())
}

func TestSSHConfig_authenticationHandlers(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	sshConfig := &SSHConfig{
		Logger: logger,
	}

	// Test password handler with nil context (should handle gracefully)
	// In real usage, context would not be nil, but we test error handling
	result := sshConfig.passwordHandler(nil, "any-password")
	assert.True(t, result)

	// Test public key handler with nil context and key (should handle gracefully)
	result = sshConfig.publicKeyHandler(nil, nil)
	assert.True(t, result)
}

// Integration test would require setting up actual SSH connections
// and Git clients, which is complex for unit tests. The above tests
// cover the core functionality of the SSH server configuration.

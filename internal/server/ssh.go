// Package server provides SSH server implementation for Git operations.
package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"path/filepath"
	"strings"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/labbs/git-server-s3/pkg/common"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
	gossh "golang.org/x/crypto/ssh"
)

// SSHConfig holds configuration for the SSH server.
// It manages Git SSH protocol operations including authentication,
// repository access, and Git command execution.
type SSHConfig struct {
	Port        string                       // SSH server port (e.g., ":2222")
	Logger      zerolog.Logger               // Logger for SSH operations
	Storage     storage.GitRepositoryStorage // Storage backend for repositories
	Server      *gliderssh.Server            // The underlying SSH server instance
	HostKeyPath string                       // Path to SSH host key file
}

// Configure sets up the SSH server with authentication handlers and Git command processing.
// It generates or loads the SSH host key and configures the server to handle Git operations.
func (sc *SSHConfig) Configure() error {
	logger := sc.Logger.With().Str("component", "ssh-server").Logger()

	// Set default host key path if not provided
	if sc.HostKeyPath == "" {
		sc.HostKeyPath = "ssh_host_key"
	}

	// Generate or load SSH host key
	signer, err := sc.ensureHostKey()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ensure host key")
		return err
	}

	// Create SSH server instance
	srv := &gliderssh.Server{
		Addr: sc.Port,
		// Authentication handlers - for production, implement proper auth
		PasswordHandler:  sc.passwordHandler,
		PublicKeyHandler: sc.publicKeyHandler,
		Handler:          sc.handleSSHSession,
	}

	// Add the host key
	srv.AddHostKey(signer)

	sc.Server = srv
	logger.Info().Str("addr", sc.Port).Msg("SSH server configured")
	return nil
}

// NewServer starts the SSH server and begins listening for connections.
// This is a blocking call that will run until the server is stopped or an error occurs.
func (sc *SSHConfig) NewServer() error {
	logger := sc.Logger.With().Str("component", "ssh-server").Logger()

	logger.Info().Str("addr", sc.Port).Msg("Starting SSH Git server")
	if err := sc.Server.ListenAndServe(); err != nil {
		logger.Error().Err(err).Msg("SSH server failed")
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the SSH server
func (sc *SSHConfig) Shutdown() error {
	if sc.Server != nil {
		sc.Logger.Info().Msg("Shutting down SSH server")
		return sc.Server.Close()
	}
	return nil
}

// ensureHostKey generates or loads an SSH host key for the server.
// If the key file doesn't exist, it generates a new ed25519 key pair.
func (sc *SSHConfig) ensureHostKey() (gossh.Signer, error) {
	logger := sc.Logger.With().Str("component", "ssh-hostkey").Logger()

	// Try to load existing key
	if data, err := os.ReadFile(sc.HostKeyPath); err == nil {
		if block, _ := pem.Decode(data); block != nil {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				if sk, ok := key.(ed25519.PrivateKey); ok {
					logger.Info().Str("path", sc.HostKeyPath).Msg("Loaded existing SSH host key")
					return gossh.NewSignerFromKey(sk)
				}
			}
		}
	}

	// Generate new key
	logger.Info().Str("path", sc.HostKeyPath).Msg("Generating new SSH host key")
	_, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Save key in PKCS#8 PEM format
	b, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(sc.HostKeyPath), 0755); err != nil {
		return nil, err
	}

	// Write key file with restrictive permissions
	f, err := os.OpenFile(sc.HostKeyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: b}); err != nil {
		return nil, err
	}

	logger.Info().Str("path", sc.HostKeyPath).Msg("SSH host key generated and saved")
	return gossh.NewSignerFromKey(sk)
}

// passwordHandler handles password-based authentication.
// In production, implement proper password validation.
func (sc *SSHConfig) passwordHandler(ctx gliderssh.Context, password string) bool {
	logger := sc.Logger.With().Str("component", "ssh-auth").Logger()

	// Handle nil context gracefully (for testing)
	if ctx != nil {
		logger = logger.With().
			Str("user", ctx.User()).
			Str("remote", ctx.RemoteAddr().String()).
			Logger()
	}

	// TODO: Implement proper password authentication
	// For now, accept any password for demonstration
	logger.Info().Msg("Password authentication accepted (demo mode)")
	return true
}

// publicKeyHandler handles public key-based authentication.
// In production, implement proper public key validation.
func (sc *SSHConfig) publicKeyHandler(ctx gliderssh.Context, key gliderssh.PublicKey) bool {
	logger := sc.Logger.With().Str("component", "ssh-auth").Logger()

	// Handle nil context gracefully (for testing)
	if ctx != nil {
		logger = logger.With().
			Str("user", ctx.User()).
			Str("remote", ctx.RemoteAddr().String()).
			Logger()
	}

	// Handle nil key gracefully (for testing)
	if key != nil {
		logger = logger.With().Str("key_type", key.Type()).Logger()
	}

	// TODO: Implement proper public key authentication
	// For now, accept any public key for demonstration
	logger.Info().Msg("Public key authentication accepted (demo mode)")
	return true
}

// handleSSHSession processes an incoming SSH session and executes Git commands.
// It parses the Git command (git-upload-pack or git-receive-pack) and delegates
// to the appropriate handler.
func (sc *SSHConfig) handleSSHSession(s gliderssh.Session) {
	logger := sc.Logger.With().
		Str("component", "ssh-session").
		Str("user", s.User()).
		Str("remote", s.RemoteAddr().String()).
		Str("command", s.RawCommand()).
		Logger()

	logger.Debug().Msg("Handling SSH session")

	// Parse the Git command
	service, repoArg := sc.parseGitCommand(s.RawCommand())
	if service == "" || repoArg == "" {
		logger.Warn().Str("command", s.RawCommand()).Msg("Invalid Git command")
		_, _ = io.WriteString(s.Stderr(), "Invalid Git command\n")
		_ = s.Exit(1)
		return
	}

	// Extract and normalize repository path
	repoPath := sc.repoPathFromSSHArg(repoArg)
	if repoPath == "" {
		logger.Warn().Str("repo_arg", repoArg).Msg("Invalid repository path")
		_, _ = io.WriteString(s.Stderr(), "Invalid repository path\n")
		_ = s.Exit(1)
		return
	}

	logger = logger.With().
		Str("service", service).
		Str("repo_path", repoPath).
		Logger()

	// Check if repository exists, create for receive-pack if needed
	exists := sc.Storage.RepositoryExists(repoPath)
	if !exists && service == "git-receive-pack" {
		logger.Info().Msg("Creating repository for push operation")
		if err := sc.Storage.CreateRepository(repoPath); err != nil {
			logger.Error().Err(err).Msg("Failed to create repository")
			_, _ = io.WriteString(s.Stderr(), "Failed to create repository: "+err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
		exists = true
	}

	if !exists {
		logger.Warn().Msg("Repository not found")
		_, _ = io.WriteString(s.Stderr(), "Repository not found\n")
		_ = s.Exit(1)
		return
	}

	// Get transport server for the repository
	srv, ep, err := common.GetTransportServer(repoPath, sc.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		_, _ = io.WriteString(s.Stderr(), "Transport server error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Handle the specific Git service
	switch service {
	case "git-upload-pack":
		sc.handleUploadPack(s, srv, ep, logger)
	case "git-receive-pack":
		sc.handleReceivePack(s, srv, ep, logger)
	default:
		logger.Error().Str("service", service).Msg("Unsupported Git service")
		_, _ = io.WriteString(s.Stderr(), "Unsupported service: "+service+"\n")
		_ = s.Exit(1)
	}
}

// handleUploadPack processes git-upload-pack requests (clone, fetch operations).
func (sc *SSHConfig) handleUploadPack(s gliderssh.Session, srv transport.Transport, ep *transport.Endpoint, logger zerolog.Logger) {
	logger.Debug().Msg("Handling upload-pack (clone/fetch)")

	// Create upload pack session
	up, err := srv.NewUploadPackSession(ep, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create upload pack session")
		_, _ = io.WriteString(s.Stderr(), "Upload pack session error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Send advertised references first (SSH protocol)
	adv, err := up.AdvertisedReferences()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get advertised references")
		_, _ = io.WriteString(s.Stderr(), "Advertised references error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	if err := adv.Encode(s); err != nil {
		logger.Error().Err(err).Msg("Failed to encode advertised references")
		_, _ = io.WriteString(s.Stderr(), "Failed to send references: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Read client request and respond
	req := packp.NewUploadPackRequest()
	if err := req.Decode(s); err != nil {
		// For empty repositories, clients may not send a proper upload pack request
		// This is normal behavior and should be handled gracefully
		if err.Error() == "pkt-line 1: missing 'want ' prefix" {
			logger.Info().Msg("Client disconnected - likely empty repository clone")
		} else {
			logger.Error().Err(err).Msg("Failed to decode upload pack request")
			_, _ = io.WriteString(s.Stderr(), "Request decode error: "+err.Error()+"\n")
		}
		_ = s.Exit(0) // Exit with success for empty repo case
		return
	}

	// Process the upload pack request
	resp, err := up.UploadPack(context.Background(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Upload pack failed")
		_, _ = io.WriteString(s.Stderr(), "Upload pack error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}
	defer resp.Close()

	// Send response to client
	if err := resp.Encode(s); err != nil {
		logger.Error().Err(err).Msg("Failed to encode upload pack response")
		_, _ = io.WriteString(s.Stderr(), "Response encode error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	logger.Info().Msg("Upload pack completed successfully")
	_ = s.Exit(0)
}

// handleReceivePack processes git-receive-pack requests (push operations).
func (sc *SSHConfig) handleReceivePack(s gliderssh.Session, srv transport.Transport, ep *transport.Endpoint, logger zerolog.Logger) {
	logger.Debug().Msg("Handling receive-pack (push)")

	// Create receive pack session
	rp, err := srv.NewReceivePackSession(ep, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create receive pack session")
		_, _ = io.WriteString(s.Stderr(), "Receive pack session error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Send advertised references first (SSH protocol)
	adv, err := rp.AdvertisedReferences()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get advertised references")
		_, _ = io.WriteString(s.Stderr(), "Advertised references error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	if err := adv.Encode(s); err != nil {
		logger.Error().Err(err).Msg("Failed to encode advertised references")
		_, _ = io.WriteString(s.Stderr(), "Failed to send references: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Read reference update request from client
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(s); err != nil {
		logger.Error().Err(err).Msg("Failed to decode receive pack request")
		_, _ = io.WriteString(s.Stderr(), "Request decode error: "+err.Error()+"\n")
		_ = s.Exit(1)
		return
	}

	// Process the receive pack request
	report, err := rp.ReceivePack(context.Background(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Receive pack failed")
		// Even on error, try to send the report to the client
		if report != nil {
			_ = report.Encode(s)
		} else {
			_, _ = io.WriteString(s.Stderr(), "Receive pack error: "+err.Error()+"\n")
		}
		_ = s.Exit(1)
		return
	}

	// Send status report to client
	if report != nil {
		if err := report.Encode(s); err != nil {
			logger.Error().Err(err).Msg("Failed to encode receive pack report")
			_, _ = io.WriteString(s.Stderr(), "Report encode error: "+err.Error()+"\n")
			_ = s.Exit(1)
			return
		}
	} else {
		// If no report, send a minimal success response
		logger.Debug().Msg("No report to send, assuming success")
	}

	logger.Info().Msg("Receive pack completed successfully")

	// Exit cleanly with success code
	_ = s.Exit(0)
}

// parseGitCommand extracts the Git service and repository argument from an SSH command.
// Example: "git-upload-pack '/demo.git'" -> ("git-upload-pack", "/demo.git")
func (sc *SSHConfig) parseGitCommand(cmd string) (service, repoArg string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "", ""
	}

	// Commands are in the form: git-upload-pack 'path' or git-receive-pack 'path'
	if strings.HasPrefix(cmd, "git-upload-pack") {
		return "git-upload-pack", strings.TrimSpace(strings.TrimPrefix(cmd, "git-upload-pack"))
	}
	if strings.HasPrefix(cmd, "git-receive-pack") {
		return "git-receive-pack", strings.TrimSpace(strings.TrimPrefix(cmd, "git-receive-pack"))
	}
	return "", ""
}

// repoPathFromSSHArg cleans the SSH repository argument and returns a normalized path.
// Example: '/demo.git' -> "demo.git"
func (sc *SSHConfig) repoPathFromSSHArg(arg string) string {
	arg = strings.TrimSpace(arg)

	// Remove quotes
	arg = strings.Trim(arg, "'\"")

	// Remove leading colon (some Git clients use it)
	arg = strings.TrimPrefix(arg, ":")

	// Remove host part if present (host:path format)
	if i := strings.Index(arg, ":"); i >= 0 {
		arg = arg[i+1:]
	}

	// Remove leading slash
	arg = strings.TrimPrefix(arg, "/")

	if arg == "" {
		return ""
	}

	// Normalize using common function
	return common.NormalizeRepoPath(arg)
}

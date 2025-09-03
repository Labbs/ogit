// Package server provides a custom SSH Git server implementation.
// This implementation uses golang.org/x/crypto/ssh directly for better control
// over the Git protocol and to avoid "remote end hung up unexpectedly" errors.
package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/labbs/git-server-s3/pkg/common"
	"github.com/labbs/git-server-s3/pkg/storage"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

// GitSSHServer represents a custom SSH server specifically designed for Git operations.
// Unlike generic SSH servers, this implementation handles the Git protocol directly.
type GitSSHServer struct {
	Port        string                       // SSH server port (e.g., ":2222")
	Logger      zerolog.Logger               // Logger for SSH operations
	Storage     storage.GitRepositoryStorage // Storage backend for repositories
	HostKeyPath string                       // Path to SSH host key file
	listener    net.Listener                 // Network listener
	sshConfig   *ssh.ServerConfig            // SSH server configuration
}

// Configure sets up the SSH server with proper Git protocol handling.
func (s *GitSSHServer) Configure() error {
	logger := s.Logger.With().Str("component", "git-ssh-server").Logger()

	// Set default host key path
	if s.HostKeyPath == "" {
		s.HostKeyPath = "ssh_host_key"
	}

	// Generate or load SSH host key
	privateKey, err := s.ensureHostKey()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to ensure host key")
		return err
	}

	// Create SSH server configuration
	s.sshConfig = &ssh.ServerConfig{
		// Demo authentication - in production, implement proper auth
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			logger.Info().
				Str("user", conn.User()).
				Str("remote", conn.RemoteAddr().String()).
				Msg("Password authentication attempt")
			// Accept any password for demo (implement proper validation in production)
			return nil, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			logger.Info().
				Str("user", conn.User()).
				Str("key_type", key.Type()).
				Str("remote", conn.RemoteAddr().String()).
				Msg("Public key authentication attempt")
			// Accept any valid key for demo (implement proper validation in production)
			return nil, nil
		},
	}

	// Add host key to server config
	s.sshConfig.AddHostKey(privateKey)

	logger.Info().Str("addr", s.Port).Msg("Git SSH server configured")
	return nil
}

// Start begins listening for SSH connections and handles Git operations.
func (s *GitSSHServer) Start() error {
	logger := s.Logger.With().Str("component", "git-ssh-server").Logger()

	// Start listening on the specified port
	listener, err := net.Listen("tcp", s.Port)
	if err != nil {
		logger.Error().Err(err).Str("addr", s.Port).Msg("Failed to listen on port")
		return err
	}
	s.listener = listener

	logger.Info().Str("addr", s.Port).Msg("Git SSH server started")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if server was closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				logger.Info().Msg("SSH server stopped")
				return nil
			}
			logger.Error().Err(err).Msg("Failed to accept connection")
			continue
		}

		// Handle connection in goroutine
		go s.handleConnection(conn)
	}
}

// Stop gracefully stops the SSH server.
func (s *GitSSHServer) Stop() error {
	s.Logger.Info().Msg("Stopping Git SSH server")
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnection processes an incoming SSH connection.
func (s *GitSSHServer) handleConnection(conn net.Conn) {
	logger := s.Logger.With().
		Str("component", "git-ssh-connection").
		Str("remote", conn.RemoteAddr().String()).
		Logger()

	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		logger.Error().Err(err).Msg("SSH handshake failed")
		return
	}
	defer sshConn.Close()

	logger.Info().Str("user", sshConn.User()).Msg("SSH connection established")

	// Handle global requests (usually none for Git)
	go ssh.DiscardRequests(reqs)

	// Handle channels (Git commands)
	for newChannel := range chans {
		go s.handleChannel(sshConn, newChannel, logger)
	}
}

// handleChannel processes SSH channels containing Git commands.
func (s *GitSSHServer) handleChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel, logger zerolog.Logger) {
	// Git operations only use "session" channel type
	if newChannel.ChannelType() != "session" {
		logger.Debug().Str("channel_type", newChannel.ChannelType()).Msg("Rejecting non-session channel")
		newChannel.Reject(ssh.UnknownChannelType, "only session channels are supported")
		return
	}

	// Accept the channel
	channel, requests, err := newChannel.Accept()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to accept channel")
		return
	}
	defer channel.Close()

	// Process channel requests
	for req := range requests {
		switch req.Type {
		case "exec":
			// Execute Git command
			s.handleExecRequest(conn, channel, req, logger)
			return // Exit after handling exec request
		default:
			// Reject other request types
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// handleExecRequest processes Git command execution requests.
func (s *GitSSHServer) handleExecRequest(conn *ssh.ServerConn, channel ssh.Channel, req *ssh.Request, logger zerolog.Logger) {
	if !req.WantReply {
		s.sendExitStatus(channel, 1)
		return
	}

	// Extract command from request payload
	command := string(req.Payload[4:]) // Skip 4-byte length prefix
	logger = logger.With().Str("command", command).Logger()

	// Parse Git command
	service, repoPath := s.parseGitCommand(command)
	if service == "" {
		logger.Error().Msg("Invalid Git command")
		req.Reply(false, nil)
		s.sendExitStatus(channel, 1)
		return
	}

	logger = logger.With().
		Str("service", service).
		Str("repo_path", repoPath).
		Logger()

	// Accept the request
	req.Reply(true, nil)

	// Handle the Git operation
	switch service {
	case "git-upload-pack":
		if err := s.handleUploadPack(channel, repoPath, logger); err != nil {
			logger.Error().Err(err).Msg("Upload pack failed")
			s.sendExitStatus(channel, 1)
		} else {
			s.sendExitStatus(channel, 0)
		}
	case "git-receive-pack":
		if err := s.handleReceivePack(channel, repoPath, logger); err != nil {
			logger.Error().Err(err).Msg("Receive pack failed")
			s.sendExitStatus(channel, 1)
		} else {
			s.sendExitStatus(channel, 0)
		}
	default:
		logger.Error().Str("service", service).Msg("Unsupported Git service")
		s.sendExitStatus(channel, 1)
	}
}

// handleUploadPack processes git-upload-pack operations (clone/fetch).
func (s *GitSSHServer) handleUploadPack(channel ssh.Channel, repoPath string, logger zerolog.Logger) error {
	logger.Info().Msg("Processing upload pack request")

	// Get transport server for the repository
	srv, endpoint, err := common.GetTransportServer(repoPath, s.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return err
	}

	// Create upload pack service
	up, err := srv.NewUploadPackSession(endpoint, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create upload pack session")
		return err
	}

	// Send advertised references
	advRefs, err := up.AdvertisedReferences()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get advertised references")
		return err
	}

	if err := advRefs.Encode(channel); err != nil {
		logger.Error().Err(err).Msg("Failed to encode advertised references")
		return err
	}

	// Read client request
	req := packp.NewUploadPackRequest()
	if err := req.Decode(channel); err != nil {
		// Handle empty repository case gracefully
		if strings.Contains(err.Error(), "missing 'want' prefix") {
			logger.Info().Msg("Client disconnected - empty repository")
			return nil
		}
		logger.Error().Err(err).Msg("Failed to decode upload pack request")
		return err
	}

	// Process upload pack
	resp, err := up.UploadPack(context.Background(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Upload pack failed")
		return err
	}
	defer resp.Close()

	// Send response to client
	if err := resp.Encode(channel); err != nil {
		logger.Error().Err(err).Msg("Failed to encode upload pack response")
		return err
	}

	logger.Info().Msg("Upload pack completed successfully")
	return nil
}

// handleReceivePack processes git-receive-pack operations (push).
func (s *GitSSHServer) handleReceivePack(channel ssh.Channel, repoPath string, logger zerolog.Logger) error {
	logger.Info().Msg("Processing receive pack request")

	// Get transport server for the repository
	srv, endpoint, err := common.GetTransportServer(repoPath, s.Storage)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get transport server")
		return err
	}

	// Create receive pack service
	rp, err := srv.NewReceivePackSession(endpoint, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create receive pack session")
		return err
	}

	// Send advertised references
	advRefs, err := rp.AdvertisedReferences()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get advertised references")
		return err
	}

	if err := advRefs.Encode(channel); err != nil {
		logger.Error().Err(err).Msg("Failed to encode advertised references")
		return err
	}

	// Read client request
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(channel); err != nil {
		logger.Error().Err(err).Msg("Failed to decode receive pack request")
		return err
	}

	// Process receive pack
	report, err := rp.ReceivePack(context.Background(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Receive pack failed")
		// Try to send error report if available
		if report != nil {
			_ = report.Encode(channel)
		}
		return err
	}

	// Send status report to client
	if report != nil {
		if err := report.Encode(channel); err != nil {
			logger.Debug().Err(err).Msg("Failed to encode receive pack report")
		} else {
			logger.Debug().Msg("Receive pack report sent successfully")
		}
	}

	logger.Info().Msg("Receive pack completed successfully")
	return nil
}

// sendExitStatus sends the exit status to the SSH client and closes the channel.
func (s *GitSSHServer) sendExitStatus(channel ssh.Channel, status int) {
	exitStatus := ssh.Marshal(struct{ Value uint32 }{uint32(status)})
	channel.SendRequest("exit-status", false, exitStatus)
}

// parseGitCommand extracts the Git service and repository path from a command.
func (s *GitSSHServer) parseGitCommand(cmd string) (service, repoPath string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "", ""
	}

	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return "", ""
	}

	service = parts[0]
	repoArg := strings.Join(parts[1:], " ")

	// Extract repository path from argument
	repoPath = s.extractRepoPath(repoArg)
	repoPath = common.NormalizeRepoPath(repoPath)

	return service, repoPath
}

// extractRepoPath extracts the repository path from the SSH argument.
func (s *GitSSHServer) extractRepoPath(arg string) string {
	// Remove quotes if present
	arg = strings.Trim(arg, "'\"")
	
	// Remove leading slash if present
	arg = strings.TrimPrefix(arg, "/")
	
	// Handle formats like "host:path" by taking only the path part
	if strings.Contains(arg, ":") {
		parts := strings.Split(arg, ":")
		if len(parts) > 1 {
			arg = parts[1]
		}
	}
	
	return arg
}

// ensureHostKey generates or loads an SSH host key.
func (s *GitSSHServer) ensureHostKey() (ssh.Signer, error) {
	logger := s.Logger.With().Str("component", "git-ssh-hostkey").Logger()

	// Try to load existing key
	if data, err := os.ReadFile(s.HostKeyPath); err == nil {
		if block, _ := pem.Decode(data); block != nil {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err == nil {
				if edKey, ok := key.(ed25519.PrivateKey); ok {
					signer, err := ssh.NewSignerFromKey(edKey)
					if err == nil {
						logger.Info().Str("path", s.HostKeyPath).Msg("Loaded existing SSH host key")
						return signer, nil
					}
				}
			}
		}
	}

	// Generate new key
	logger.Info().Str("path", s.HostKeyPath).Msg("Generating new SSH host key")
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ED25519 key: %w", err)
	}

	// Convert to PKCS8 format
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.HostKeyPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write key to file
	if err := os.WriteFile(s.HostKeyPath, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		return nil, fmt.Errorf("failed to write host key: %w", err)
	}

	// Create signer
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	logger.Info().Str("path", s.HostKeyPath).Msg("SSH host key generated and saved")
	return signer, nil
}

# SSH Server for Git Operations

This package provides a complete SSH server implementation for Git operations, supporting both `git-upload-pack` (clone/fetch) and `git-receive-pack` (push) operations over SSH.

## Features

- **Full Git SSH Protocol Support**: Implements the complete Git SSH transport protocol
- **Storage Backend Agnostic**: Works with any storage backend that implements `GitRepositoryStorage`
- **Automatic Repository Creation**: Creates repositories automatically on first push
- **Comprehensive Logging**: Detailed logging for all SSH operations and Git commands
- **Security Ready**: Framework for implementing proper authentication (passwords, SSH keys)
- **Host Key Management**: Automatic generation and management of SSH host keys

## Architecture

The SSH server consists of several key components:

### SSHConfig
Main configuration struct that holds:
- SSH server port and settings
- Logger for operation tracking
- Storage backend interface
- SSH server instance

### Authentication Handlers
- `passwordHandler`: Handles password-based authentication
- `publicKeyHandler`: Handles SSH key-based authentication
- Currently configured for demo mode (accepts any credentials)

### Git Command Processing
- Parses incoming SSH commands (`git-upload-pack`, `git-receive-pack`)
- Validates repository paths and permissions
- Delegates to appropriate Git protocol handlers

### Protocol Handlers
- `handleUploadPack`: Processes clone/fetch operations
- `handleReceivePack`: Processes push operations
- Full implementation of Git wire protocol over SSH

## Usage

### Basic Setup

```go
import (
    "github.com/labbs/git-server-s3/internal/server"
    "github.com/labbs/git-server-s3/pkg/storage/local"
    "github.com/rs/zerolog"
)

// Setup storage backend
localStorage := local.NewLocalStorage(logger)
localStorage.Configure()

// Create SSH server
sshConfig := &server.SSHConfig{
    Port:        ":2222",
    Logger:      logger,
    Storage:     localStorage,
    HostKeyPath: "./ssh_host_key",
}

// Configure and start
sshConfig.Configure()
sshConfig.NewServer() // Blocking call
```

### Configuration Options

```go
type SSHConfig struct {
    Port        string                        // SSH port (e.g., ":2222")
    Logger      zerolog.Logger                // Logger instance
    Storage     storage.GitRepositoryStorage  // Storage backend
    Server      *gliderssh.Server            // SSH server instance
    HostKeyPath string                       // SSH host key file path
}
```

## Git Commands Supported

### Clone/Fetch Operations
```bash
git clone ssh://git@localhost:2222/repo.git
git fetch origin
git pull origin main
```

### Push Operations
```bash
git push origin main
git push origin --all
git push origin --tags
```

### Repository Management
- Repositories are created automatically on first push
- Repository names are normalized (e.g., `repo` becomes `repo.git`)
- Repository existence is checked before clone/fetch operations

## Security Considerations

### Current Implementation (Demo Mode)
- Accepts any password
- Accepts any SSH public key
- No user isolation
- No repository access controls

### Production Recommendations

#### Authentication
```go
// Implement proper password validation
func (sc *SSHConfig) passwordHandler(ctx gliderssh.Context, password string) bool {
    user := ctx.User()
    // Implement your password validation logic
    return validateUserPassword(user, password)
}

// Implement proper public key validation
func (sc *SSHConfig) publicKeyHandler(ctx gliderssh.Context, key gliderssh.PublicKey) bool {
    user := ctx.User()
    // Implement your public key validation logic
    return validateUserPublicKey(user, key)
}
```

#### Access Control
- Implement repository-level permissions
- Add user-based access controls
- Consider implementing organization/team permissions

#### Security Headers
- Configure proper SSH cipher suites
- Implement rate limiting
- Add IP-based restrictions if needed

## Integration with Storage Backends

The SSH server works with any storage backend implementing `GitRepositoryStorage`:

### Local Storage
```go
localStorage := local.NewLocalStorage(logger)
localStorage.Configure()
```

### S3 Storage (when implemented)
```go
s3Storage := s3.NewS3Storage(logger)
s3Storage.Configure()
```

## Logging

The SSH server provides comprehensive logging:

```json
{
  "level": "info",
  "component": "ssh-session",
  "user": "git",
  "remote": "127.0.0.1:54321",
  "command": "git-upload-pack '/repo.git'",
  "service": "git-upload-pack",
  "repo_path": "repo.git",
  "message": "Upload pack completed successfully"
}
```

## Testing

Run tests with:
```bash
go test ./internal/server/...
```

Test coverage includes:
- SSH server configuration
- Git command parsing
- Repository path normalization  
- Host key generation and loading
- Authentication handler behavior

## Examples

See `examples/ssh_server_example.go` for a complete working example.

## Dependencies

- `github.com/gliderlabs/ssh`: SSH server implementation
- `github.com/go-git/go-git/v5`: Git protocol implementation
- `golang.org/x/crypto/ssh`: SSH cryptography
- `github.com/rs/zerolog`: Structured logging

## Troubleshooting

### Common Issues

1. **Port already in use**: Make sure the SSH port isn't being used by another service
2. **Permission denied**: Check that the host key file has proper permissions (600)
3. **Repository not found**: Ensure the repository name includes `.git` suffix
4. **Authentication failures**: Verify authentication handlers are properly configured

### Debug Logging

Enable debug logging to troubleshoot issues:
```go
logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
```

This will show detailed information about:
- SSH connection establishment
- Git command parsing
- Repository operations
- Protocol message exchange

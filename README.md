# oGit

A lightweight Git server supporting both HTTP and SSH protocols with multiple storage backends.

## Features

### Current Features ✅

- **HTTP Git Server**: Full Git Smart HTTP protocol support
  - Clone, push, pull operations via HTTP/HTTPS
  - REST API for repository management
  - Graceful shutdown handling

- **SSH Git Server**: Custom SSH implementation for Git operations
  - Clone, push, pull operations via SSH
  - Public key and password authentication
  - Git protocol over SSH (git-upload-pack, git-receive-pack)

- **Storage Backends**:
  - **Local**: File system storage for repositories
  - **S3**: Amazon S3 compatible storage (tested and working)

- **Configuration**:
  - YAML configuration files
  - Command-line flags
  - Environment variable support

- **Authentication**:
  - Demo mode (password: "demo", accepts any SSH key)
  - Extensible authentication framework

- **Logging**: Structured logging with zerolog

### Architecture

- **Parallel Servers**: HTTP and SSH servers run concurrently
- **Graceful Shutdown**: CTRL+C handling with 30-second timeout
- **Storage Abstraction**: Pluggable storage backend system
- **Transport Abstraction**: Unified Git transport layer

## Quick Start

### Configuration

Copy the example configuration:
```bash
cp config-example.yaml config.yaml
```

Edit `config.yaml` to configure your storage backend and server settings.

### Running the Server

```bash
# Build the application
make build

# Run with default configuration
./main

# Or run with custom config
./main --config ./config.yaml
```

### Git Operations

**HTTP Access:**
```bash
# Clone via HTTP
git clone http://localhost:8080/your-repo.git

# Push to HTTP
git push origin main
```

**SSH Access:**
**Not stable**
```bash
# Clone via SSH (demo password: "demo")
git clone ssh://demo@localhost:2022/your-repo.git

# Push via SSH
git push origin main
```

## Configuration Options

### Server Configuration
- `server.http.enabled`: Enable/disable HTTP server
- `server.http.port`: HTTP server port (default: 8080)
- `server.ssh.enabled`: Enable/disable SSH server  
- `server.ssh.port`: SSH server port (default: 2022)
- `server.ssh.hostkey`: Path to SSH host key file

### Storage Configuration
- `storage.type`: Storage backend ("local" or "s3")
- `storage.local.path`: Local storage directory
- `storage.s3.*`: S3 configuration options

## Known Issues 🐛

### SSH Protocol
- **Client Warning**: Git clients may show "remote end hung up unexpectedly" message
  - **Impact**: Cosmetic only - all operations complete successfully
  - **Status**: Under investigation - server-side operations work perfectly
  - **Workaround**: Message can be safely ignored

### Performance
- **S3 Storage**: Slower than local storage due to network latency
  - **Expected**: Normal behavior for remote storage
  - **Optimization**: Consider using S3 transfer acceleration

## Future Features 🚀

### Authentication & Security
- [ ] Multi-user authentication system
- [ ] Role-based access control (RBAC)
- [ ] JWT token authentication
- [ ] LDAP/Active Directory integration
- [ ] Rate limiting and DDoS protection

### Storage Enhancements  
- [ ] Azure Blob Storage backend
- [ ] Google Cloud Storage backend
- [ ] Redis caching layer
- [ ] Repository compression and deduplication

### Git Features
- [ ] Web UI for repository browsing
- [ ] Webhook support for CI/CD integration
- [ ] Branch protection rules
- [ ] Repository mirroring
- [ ] Git LFS (Large File Storage) support

### Operations & Monitoring
- [ ] Metrics and monitoring (Prometheus)
- [ ] Health check endpoints
- [ ] Docker containerization
- [ ] Kubernetes deployment manifests
- [ ] Backup and restore tools

### Protocol Improvements
- [ ] Git protocol v2 support
- [ ] HTTP/2 support
- [ ] TLS certificate management
- [ ] SSH key management interface

## Development

### Building
```bash
make build
```

### Testing
```bash
make test
```

### Running Tests with Coverage
```bash
make test-coverage
```

## License

See [LICENSE](LICENSE) file for details.

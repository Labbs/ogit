# SSH Git Server

## WARNING

The SSH support is not stable and have issues (hung up error message from git cli).
Don't use this option for the moment and prefer HTTP.

This implementation provides SSH support for Git operations alongside the existing HTTP server.

## Features

- **Parallel Server Startup**: Both HTTP and SSH servers run concurrently
- **Graceful Shutdown**: CTRL+C stops both servers cleanly
- **Git Protocol Support**: Full support for `git clone`, `git push`, and `git pull` over SSH
- **Authentication**: Supports both password and public key authentication (demo mode)
- **Host Key Management**: Automatic SSH host key generation and persistence

## Usage

### Command Line

Start both HTTP and SSH servers:

```bash
./git-server-s3 server \
  --ssh.enabled=true \
  --ssh.port=2222 \
  --ssh.hostkey=./ssh_host_key \
  --http.port=8080 \
  --logger.level=debug \
  --storage.type=local \
  --storage.local.path=./repositories
```

### Configuration File

Create a `config.yaml` file:

```yaml
http:
  port: 8080
  logs: true
ssh:
  enabled: true
  port: 2222
  hostkey: ./ssh_host_key
logger:
  level: debug
  pretty: true
storage:
  type: local
  local:
    path: ./repositories
```

Then run:
```bash
./git-server-s3 server --config config.yaml
```

### Environment Variables

```bash
export SSH_ENABLED=true
export SSH_PORT=2222
export SSH_HOST_KEY_PATH=./ssh_host_key
export HTTP_PORT=8080
./git-server-s3 server
```

## Git Operations

Once the server is running, you can use Git over SSH:

### Clone a repository
```bash
git clone ssh://git@localhost:2222/my-repo.git
```

### Push to a repository
```bash
cd my-repo
git remote set-url origin ssh://git@localhost:2222/my-repo.git
git push origin main
```

### Pull from a repository
```bash
git pull origin main
```

## Authentication

The current implementation includes demo authentication handlers:

- **Password Authentication**: Any username with password "demo" is accepted
- **Public Key Authentication**: Any valid SSH public key is accepted

⚠️ **Security Note**: This is for demonstration purposes only. In production, implement proper authentication against your user database.

## Host Key Management

- SSH host key is automatically generated on first run
- Key is saved to the specified path (default: `./ssh_host_key`)
- Key is reused on subsequent starts for client trust
- Uses Ed25519 key type for better security and performance

## Graceful Shutdown

The servers support graceful shutdown:

- Press `CTRL+C` or send `SIGINT`/`SIGTERM` to stop both servers
- Servers complete ongoing requests before shutting down
- 30-second timeout for forced shutdown if needed

## Logging

All SSH operations are logged with structured logging:

- Connection events
- Authentication attempts
- Git command execution
- Error conditions

Example log output:
```
2:34PM INF Starting HTTP server port=8080
2:34PM INF Starting SSH server port=2222
2:34PM INF SSH Git server started addr=:2222
^C2:34PM INF Shutdown signal received, stopping servers...
2:34PM INF Shutting down HTTP server
2:34PM INF Shutting down SSH server
2:34PM INF All servers stopped gracefully
```

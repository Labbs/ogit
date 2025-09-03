package config

var (
	// Version is set during the startup process.
	Version string

	// ConfigFile is the path to the configuration file.
	// This is used to load the configuration at startup.
	ConfigFile string

	// Server is the configuration for the HTTP fiber server.
	// Port is the port on which the server listens.
	// HttpLogs enables or disables HTTP request logging.
	Server struct {
		Port     int
		HttpLogs bool
	}

	// SSH is the configuration for the SSH Git server.
	// Enabled controls whether the SSH server starts.
	// Port is the port on which the SSH server listens.
	// HostKeyPath is the path to the SSH host key file.
	SSH struct {
		Enabled     bool
		Port        int
		HostKeyPath string
	}

	// Logger is the configuration for the zerolog logger.
	// Level is the log level for the logger.
	// Pretty enables or disables pretty printing of logs (non JSON logs).
	Logger struct {
		Level  string
		Pretty bool
	}

	// StorageType is the type of storage to use (e.g., local, s3).
	Storage struct {
		Type string

		S3 struct {
			Bucket    string
			Endpoint  string
			AccessKey string
			SecretKey string
			Region    string
		}

		Local struct {
			Path string
		}
	}
)

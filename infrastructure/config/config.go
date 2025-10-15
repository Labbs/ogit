package config

type Config struct {
	// Version of the application
	Version string

	// ConfigFile is the path to the configuration file
	ConfigFile string

	Server struct {
		// Port is the server port
		Port int
		// HttpLogs indicates if HTTP logs are enabled
		HttpLogs bool
	}

	// Logger is the configuration for the zerolog logger.
	// Level is the log level for the logger.
	// Pretty enables or disables pretty printing of logs (non JSON logs).
	Logger struct {
		Level  string
		Pretty bool
	}

	// Database is the configuration for the database connection.
	// Dialect is the database engine (sqlite, postgres, etc.).
	// DSN is the Data Source Name for the database connection.
	Database struct {
		Dialect string // Database engine (sqlite, postgres, etc.)
		DSN     string
	}

	Session struct {
		SecretKey         string
		ExpirationMinutes int
		Issuer            string
	}

	Auth struct {
		DisableAdminAccount bool
	}

	Registration struct {
		Enabled                  bool     // Enable or disable user registration
		RequireEmailVerification bool     // Require email verification for new registrations
		DomainWhitelist          []string // List of allowed domains for registration
		PasswordMinLength        int      // Minimum password length for registration
		PasswordComplexity       bool     // Require complex passwords (uppercase, lowercase, numbers, symbols)
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
}

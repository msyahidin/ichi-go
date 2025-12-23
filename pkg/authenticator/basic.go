package authenticator

func NewBasicAuthenticator(config *BasicAuthConfig) *BasicAuthenticator {
	return &BasicAuthenticator{config: config}
}

type BasicAuthenticator struct {
	config *BasicAuthConfig
}

type BasicAuthConfig struct {
	Enabled bool

	// Credential validation
	Validator BasicAuthValidator // Custom username/password validator
	Realm     string             // WWW-Authenticate realm (default: "Restricted")

	// For simple static credentials (dev/testing only)
	Users map[string]string // username -> password (hashed!)

	// Advanced
	SkipPaths []string
}

type BasicAuthValidator func(username, password string) (userID string, err error)

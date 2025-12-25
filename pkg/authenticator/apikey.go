package authenticator

func NewAPIKeyAuthenticator(config *APIKeyConfig) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{config: config}
}

type APIKeyAuthenticator struct {
	config *APIKeyConfig
}

type APIKeyConfig struct {
	Enabled   bool
	Header    string
	Validator APIKeyValidator
	SkipPaths []string
}

type APIKeyValidator func(apiKey string) (userID string, err error)

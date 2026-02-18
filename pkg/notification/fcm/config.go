package fcm

import "github.com/spf13/viper"

// Config holds Firebase Cloud Messaging configuration.
type Config struct {
	// CredentialsFile is the path to the Firebase service account JSON key file.
	// Leave empty when running on GCP — Application Default Credentials (ADC) are used automatically.
	// NEVER commit the service account file to source control.
	CredentialsFile string `mapstructure:"credentials_file"`

	// ProjectID is the Firebase/GCP project ID.
	// Used for logging and validation; not strictly required for sending if credentials are valid.
	ProjectID string `mapstructure:"project_id"`

	// Enabled controls whether the FCM client is initialized.
	// Set to false to skip FCM initialization in environments without credentials.
	Enabled bool `mapstructure:"enabled"`
}

// SetDefault registers Viper defaults for the FCM config block.
// Called from config.setDefault() during application startup.
func SetDefault() {
	viper.SetDefault("notification.fcm.enabled", false)
	viper.SetDefault("notification.fcm.credentials_file", "")
	viper.SetDefault("notification.fcm.project_id", "")
}

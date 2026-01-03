package errors

var serviceName = "ichi-go"

// SetServiceName sets the service name for error context
func SetServiceName(name string) {
	serviceName = name
}

// GetServiceName returns the current service name
func GetServiceName() string {
	return serviceName
}

package river

import (
	"fmt"

	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
)

// RegisterBridgeWorkers builds a single BridgeWorker from all ConsumerRegistrations
// and adds it once. Adding one worker per registration would panic because they all
// share the same Kind() == "generic_job".
// Returns an error if duplicate ConsumerRegistration names are detected.
func RegisterBridgeWorkers(workers *riverqueue.Workers, registrations []queue.ConsumerRegistration) error {
	handlers := make(map[string]queue.ConsumeFunc, len(registrations))
	for _, reg := range registrations {
		if _, exists := handlers[reg.Name]; exists {
			return fmt.Errorf("river: duplicate consumer registration for name %q", reg.Name)
		}
		handlers[reg.Name] = reg.ConsumeFunc
	}
	riverqueue.AddWorker(workers, NewBridgeWorker(handlers))
	return nil
}

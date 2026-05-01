package river

import (
	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
)

// RegisterBridgeWorkers builds a single BridgeWorker from all ConsumerRegistrations
// and adds it once. Adding one worker per registration would panic because they all
// share the same Kind() == "generic_job".
func RegisterBridgeWorkers(workers *riverqueue.Workers, registrations []queue.ConsumerRegistration) {
	handlers := make(map[string]queue.ConsumeFunc, len(registrations))
	for _, reg := range registrations {
		handlers[reg.Name] = reg.ConsumeFunc
	}
	riverqueue.AddWorker(workers, NewBridgeWorker(handlers))
}

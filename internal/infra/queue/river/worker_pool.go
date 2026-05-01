package river

import (
	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
)

// RegisterBridgeWorkers adds a GenericJobWorker for each ConsumerRegistration.
// Called during River client setup so existing consumers work unchanged with the River driver.
func RegisterBridgeWorkers(workers *riverqueue.Workers, registrations []queue.ConsumerRegistration) {
	for _, reg := range registrations {
		riverqueue.AddWorker(workers, NewGenericJobWorker(reg.ConsumeFunc))
	}
}

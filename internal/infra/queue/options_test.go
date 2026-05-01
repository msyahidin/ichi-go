package queue_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ichi-go/internal/infra/queue"
)

func TestDispatchOptions_Defaults(t *testing.T) {
	o := queue.ApplyOptions()
	assert.Equal(t, "default", o.Queue)
	assert.Equal(t, 3, o.MaxAttempts)
	assert.Equal(t, time.Duration(0), o.Delay)
	assert.Equal(t, 1, o.Priority)
}

func TestDispatchOptions_WithOptions(t *testing.T) {
	o := queue.ApplyOptions(
		queue.OnQueue("emails"),
		queue.Delay(5*time.Minute),
		queue.MaxAttempts(5),
		queue.Priority(2),
	)
	assert.Equal(t, "emails", o.Queue)
	assert.Equal(t, 5*time.Minute, o.Delay)
	assert.Equal(t, 5, o.MaxAttempts)
	assert.Equal(t, 2, o.Priority)
}

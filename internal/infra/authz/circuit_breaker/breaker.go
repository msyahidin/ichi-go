package circuit_breaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"ichi-go/pkg/logger"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed - Normal operation, requests pass through
	StateClosed State = iota
	// StateOpen - Circuit is open, requests fail fast
	StateOpen
	// StateHalfOpen - Testing if service recovered
	StateHalfOpen
)

var (
	// ErrCircuitOpen is returned when the circuit is open
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests is returned when max requests in half-open state exceeded
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	state         State
	failureCount  int
	successCount  int
	lastFailTime  time.Time
	lastStateTime time.Time

	// Configuration
	maxFailures       int           // Failures before opening
	timeout           time.Duration // Time to wait in open state before half-open
	halfOpenMaxReqs   int           // Max requests allowed in half-open state
	halfOpenSuccesses int           // Successes needed to close from half-open

	mu sync.RWMutex

	// Metrics
	totalRequests  int64
	totalSuccesses int64
	totalFailures  int64
	totalRejects   int64
}

// Config configures the circuit breaker
type Config struct {
	Name              string        // Circuit breaker name (for logging)
	MaxFailures       int           // Failures before opening (default: 5)
	Timeout           time.Duration // Time in open state (default: 60s)
	HalfOpenMaxReqs   int           // Max requests in half-open (default: 3)
	HalfOpenSuccesses int           // Successes to close (default: 2)
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		Name:              "rbac",
		MaxFailures:       5,
		Timeout:           60 * time.Second,
		HalfOpenMaxReqs:   3,
		HalfOpenSuccesses: 2,
	}
}

// New creates a new circuit breaker
func New(config Config) *CircuitBreaker {
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.HalfOpenMaxReqs == 0 {
		config.HalfOpenMaxReqs = 3
	}
	if config.HalfOpenSuccesses == 0 {
		config.HalfOpenSuccesses = 2
	}

	return &CircuitBreaker{
		name:              config.Name,
		state:             StateClosed,
		maxFailures:       config.MaxFailures,
		timeout:           config.Timeout,
		halfOpenMaxReqs:   config.HalfOpenMaxReqs,
		halfOpenSuccesses: config.HalfOpenSuccesses,
		lastStateTime:     time.Now(),
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if allowed to execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute function
	err := fn()

	// Record result
	cb.afterRequest(err)

	return err
}

// ExecuteWithFallback executes a function with fallback on circuit open
func (cb *CircuitBreaker) ExecuteWithFallback(
	ctx context.Context,
	fn func() error,
	fallback func() error,
) error {
	err := cb.Execute(ctx, fn)

	// If circuit is open, use fallback
	if errors.Is(err, ErrCircuitOpen) {
		logger.WithContext(ctx).Warnf(
			"Circuit %s is open, using fallback",
			cb.name,
		)
		return fallback()
	}

	return err
}

// beforeRequest checks if request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++

	switch cb.state {
	case StateClosed:
		// Normal operation, allow request
		return nil

	case StateOpen:
		// Check if timeout expired
		if time.Since(cb.lastStateTime) > cb.timeout {
			// Transition to half-open
			cb.setState(StateHalfOpen)
			logger.Infof("Circuit %s transitioned to half-open", cb.name)
			return nil
		}

		// Circuit still open, reject
		cb.totalRejects++
		return ErrCircuitOpen

	case StateHalfOpen:
		// Check if we've exceeded max requests in half-open
		if cb.successCount+cb.failureCount >= cb.halfOpenMaxReqs {
			cb.totalRejects++
			return ErrTooManyRequests
		}
		return nil

	default:
		return nil
	}
}

// afterRequest records the result and updates state
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.totalSuccesses++

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0

	case StateHalfOpen:
		cb.successCount++

		// Check if enough successes to close circuit
		if cb.successCount >= cb.halfOpenSuccesses {
			cb.setState(StateClosed)
			cb.failureCount = 0
			cb.successCount = 0
			logger.Infof("Circuit %s closed after successful recovery", cb.name)
		}
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure() {
	cb.totalFailures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++

		// Check if threshold reached
		if cb.failureCount >= cb.maxFailures {
			cb.setState(StateOpen)
			logger.Warnf(
				"Circuit %s opened after %d failures",
				cb.name, cb.failureCount,
			)
		}

	case StateHalfOpen:
		// Any failure in half-open state reopens circuit
		cb.setState(StateOpen)
		cb.successCount = 0
		cb.failureCount = 0
		logger.Warnf("Circuit %s reopened during recovery test", cb.name)
	}
}

// setState transitions to a new state
func (cb *CircuitBreaker) setState(state State) {
	if cb.state != state {
		cb.state = state
		cb.lastStateTime = time.Now()
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// IsOpen returns true if circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == StateOpen
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateTime = time.Now()

	logger.Infof("Circuit %s manually reset", cb.name)
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() Metrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	successRate := float64(0)
	if cb.totalRequests > 0 {
		successRate = float64(cb.totalSuccesses) / float64(cb.totalRequests)
	}

	return Metrics{
		Name:           cb.name,
		State:          cb.state,
		TotalRequests:  cb.totalRequests,
		TotalSuccesses: cb.totalSuccesses,
		TotalFailures:  cb.totalFailures,
		TotalRejects:   cb.totalRejects,
		SuccessRate:    successRate,
		FailureCount:   cb.failureCount,
		LastStateTime:  cb.lastStateTime,
	}
}

// Metrics represents circuit breaker metrics
type Metrics struct {
	Name           string
	State          State
	TotalRequests  int64
	TotalSuccesses int64
	TotalFailures  int64
	TotalRejects   int64
	SuccessRate    float64
	FailureCount   int
	LastStateTime  time.Time
}

// String returns a string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// String returns a string representation of metrics
func (m Metrics) String() string {
	return fmt.Sprintf(
		"Circuit[%s] state=%s requests=%d successes=%d failures=%d rejects=%d rate=%.2f%%",
		m.Name, m.State, m.TotalRequests, m.TotalSuccesses,
		m.TotalFailures, m.TotalRejects, m.SuccessRate*100,
	)
}

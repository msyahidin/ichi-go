package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ichi-go/internal/infra/authz/cache"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

// RBACEvent represents an RBAC change event for cache invalidation
type RBACEvent struct {
	EventID   string       `json:"event_id"`
	Timestamp time.Time    `json:"timestamp"`
	Action    string       `json:"action"` // policy_added, policy_removed, role_assigned, role_revoked
	TenantID  string       `json:"tenant_id"`
	SubjectID string       `json:"subject_id"` // User affected
	Details   EventDetails `json:"details"`
}

// EventDetails contains the specifics of the RBAC change
type EventDetails struct {
	Role         string   `json:"role,omitempty"`
	Resource     string   `json:"resource,omitempty"`
	Action       string   `json:"action,omitempty"`
	CacheKeys    []string `json:"cache_keys,omitempty"`    // Specific keys to invalidate
	ReloadPolicy bool     `json:"reload_policy,omitempty"` // Whether to reload policies
}

// RabbitMQWatcher listens to RBAC events and invalidates cache
type RabbitMQWatcher struct {
	consumer      rabbitmq.MessageConsumer
	decisionCache *cache.DecisionCache
	ctx           context.Context
	cancel        context.CancelFunc
}

// WatcherConfig configures the RabbitMQ watcher
type WatcherConfig struct {
	ExchangeName string
	QueueName    string
	RoutingKeys  []string
	ConsumerTag  string
}

// NewRabbitMQWatcher creates a new RabbitMQ watcher for RBAC events
func NewRabbitMQWatcher(
	consumer rabbitmq.MessageConsumer,
	decisionCache *cache.DecisionCache,
	config WatcherConfig,
) (*RabbitMQWatcher, error) {
	if consumer == nil {
		return nil, fmt.Errorf("RabbitMQ consumer is required")
	}
	if decisionCache == nil {
		return nil, fmt.Errorf("decision cache is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	watcher := &RabbitMQWatcher{
		consumer:      consumer,
		decisionCache: decisionCache,
		ctx:           ctx,
		cancel:        cancel,
	}

	logger.Infof("RabbitMQ RBAC watcher initialized for queue: %s", config.QueueName)

	return watcher, nil
}

// Start begins listening to RBAC events
func (w *RabbitMQWatcher) Start() error {
	logger.Infof("Starting RabbitMQ RBAC watcher")

	// Start consuming messages with handler
	go func() {
		if err := w.consumer.Consume(w.ctx, w.handleMessage); err != nil {
			logger.Errorf("RBAC watcher consume error: %v", err)
		}
	}()

	return nil
}

// Stop stops the watcher gracefully
func (w *RabbitMQWatcher) Stop() error {
	logger.Infof("Stopping RabbitMQ RBAC watcher")

	w.cancel()

	// Close consumer
	if w.consumer != nil {
		w.consumer.Close()
	}

	return nil
}

// handleMessage processes an RBAC event message (ConsumeFunc)
func (w *RabbitMQWatcher) handleMessage(ctx context.Context, body []byte) error {
	var event RBACEvent

	// Parse event
	if err := json.Unmarshal(body, &event); err != nil {
		logger.WithContext(ctx).Errorf("Failed to parse RBAC event: %v", err)
		return nil // Don't retry bad JSON (permanent failure)
	}

	logger.WithContext(ctx).Infof(
		"Received RBAC event: action=%s tenant=%s subject=%s",
		event.Action, event.TenantID, event.SubjectID,
	)

	// Process event based on action
	if err := w.processEvent(ctx, &event); err != nil {
		logger.WithContext(ctx).Errorf("Failed to process RBAC event: %v", err)
		return err // Retry on processing error (transient failure)
	}

	return nil
}

// processEvent handles cache invalidation based on the event
func (w *RabbitMQWatcher) processEvent(ctx context.Context, event *RBACEvent) error {
	switch event.Action {
	case "policy_added", "policy_removed":
		return w.handlePolicyChange(ctx, event)

	case "role_assigned", "role_revoked":
		return w.handleRoleChange(ctx, event)

	case "permission_granted", "permission_revoked":
		return w.handlePermissionChange(ctx, event)

	default:
		logger.WithContext(ctx).Warnf("Unknown RBAC event action: %s", event.Action)
		return nil
	}
}

// handlePolicyChange invalidates cache when policies change
func (w *RabbitMQWatcher) handlePolicyChange(ctx context.Context, event *RBACEvent) error {
	// Invalidate all decision cache for the tenant
	pattern := cache.MakeTenantPattern(event.TenantID)

	if err := w.decisionCache.DeletePattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate tenant cache: %w", err)
	}

	logger.WithContext(ctx).Infof(
		"Invalidated cache for tenant %s due to policy change",
		event.TenantID,
	)

	return nil
}

// handleRoleChange invalidates cache when user roles change
func (w *RabbitMQWatcher) handleRoleChange(ctx context.Context, event *RBACEvent) error {
	// Invalidate cache for specific user in tenant
	pattern := cache.MakeUserPattern(event.TenantID, event.SubjectID)

	if err := w.decisionCache.DeletePattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	logger.WithContext(ctx).Infof(
		"Invalidated cache for user %s in tenant %s due to role change",
		event.SubjectID, event.TenantID,
	)

	return nil
}

// handlePermissionChange invalidates cache when specific permissions change
func (w *RabbitMQWatcher) handlePermissionChange(ctx context.Context, event *RBACEvent) error {
	// If specific cache keys provided, delete those
	if len(event.Details.CacheKeys) > 0 {
		for _, key := range event.Details.CacheKeys {
			if err := w.decisionCache.Delete(ctx, key); err != nil {
				logger.WithContext(ctx).Errorf("Failed to delete cache key %s: %v", key, err)
			}
		}
		return nil
	}

	// Otherwise, invalidate user's cache
	return w.handleRoleChange(ctx, event)
}

// PublishEvent publishes an RBAC event for cache invalidation
// This is a helper function for publishing events from the RBAC service
func PublishEvent(
	ctx context.Context,
	publisher rabbitmq.MessageProducer,
	event *RBACEvent,
) error {
	// Set event metadata
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("rbac_%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Determine routing key
	routingKey := fmt.Sprintf("rbac.%s.%s", event.Action, event.TenantID)

	// Publish event using internal interface
	err := publisher.Publish(
		ctx,
		routingKey,
		event,
		rabbitmq.PublishOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to publish RBAC event: %w", err)
	}

	logger.WithContext(ctx).Debugf("Published RBAC event: %s to %s", event.EventID, routingKey)

	return nil
}

// CreateRBACExchange creates the RBAC events exchange configuration
func CreateRBACExchange() map[string]interface{} {
	return map[string]interface{}{
		"name":        "rbac.events",
		"type":        "topic",
		"durable":     true,
		"auto_delete": false,
		"internal":    false,
		"no_wait":     false,
	}
}

// CreateRBACQueue creates the RBAC cache invalidation queue configuration
func CreateRBACQueue() map[string]interface{} {
	return map[string]interface{}{
		"name":        "rbac.cache.invalidation",
		"durable":     true,
		"auto_delete": false,
		"exclusive":   false,
		"no_wait":     false,
	}
}

// GetRBACRoutingKeys returns the routing keys for RBAC events
func GetRBACRoutingKeys() []string {
	return []string{
		"rbac.policy_added.*",
		"rbac.policy_removed.*",
		"rbac.role_assigned.*",
		"rbac.role_revoked.*",
		"rbac.permission_granted.*",
		"rbac.permission_revoked.*",
	}
}

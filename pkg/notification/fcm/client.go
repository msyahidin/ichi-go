package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client wraps the Firebase Admin SDK messaging client with convenience methods
// for the notification domain.
//
// Pitfalls to be aware of:
//   - Unregistered tokens (ErrCodeUnregistered): the token is permanently invalid.
//     Do NOT requeue the message. Mark the token as stale in your token store.
//   - Batch limit: SendEach accepts max 500 messages per call.
//     Use SendToTokensBatch for slices larger than 500.
//   - ADC on GCP: when CredentialsFile is empty, Firebase SDK uses Application
//     Default Credentials automatically. No code change needed.
type Client struct {
	messaging *messaging.Client
}

// NewClient initializes a Firebase app and returns an FCM messaging client.
//
//   - credentialsFile: path to the service account JSON key file.
//     Pass empty string on GCP to use Application Default Credentials.
func NewClient(ctx context.Context, credentialsFile string) (*Client, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to initialize firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to get messaging client: %w", err)
	}

	return &Client{messaging: msgClient}, nil
}

// SendToToken sends a push notification to a single device token.
//
// Returns nil for unregistered tokens — the caller should detect
// messaging.ErrCodeUnregistered in the underlying error and handle token cleanup.
func (c *Client) SendToToken(ctx context.Context, token, title, body string, data map[string]string) error {
	_, err := c.messaging.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	})
	return err
}

// SendToTokens sends a notification to multiple device tokens in a single batch call.
// FCM's SendEach processes each token independently — partial success is possible.
//
// Returns:
//   - failedTokens: tokens that failed delivery (may include stale/unregistered tokens)
//   - err: non-nil only when the entire batch request failed (network error, auth error)
//
// NOTE: FCM limits SendEach to 500 messages per call. Use SendToTokensBatch for
// slices larger than 500.
func (c *Client) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) (failedTokens []string, err error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	if len(tokens) > 500 {
		return nil, fmt.Errorf("fcm: SendToTokens accepts max 500 tokens per call, got %d; use SendToTokensBatch", len(tokens))
	}

	messages := make([]*messaging.Message, len(tokens))
	for i, token := range tokens {
		messages[i] = &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
		}
	}

	// SendEach replaces the deprecated SendAll — each message is sent independently.
	br, err := c.messaging.SendEach(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("fcm: batch send failed: %w", err)
	}

	for i, resp := range br.Responses {
		if !resp.Success {
			failedTokens = append(failedTokens, tokens[i])
		}
	}
	return failedTokens, nil
}

// SendToTokensBatch handles slices of any size by chunking into 500-token batches.
// Use this for user notification lists that may exceed 500 device tokens.
func (c *Client) SendToTokensBatch(ctx context.Context, tokens []string, title, body string, data map[string]string) (failedTokens []string, err error) {
	const batchSize = 500
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]
		failed, batchErr := c.SendToTokens(ctx, batch, title, body, data)
		if batchErr != nil {
			return failedTokens, batchErr
		}
		failedTokens = append(failedTokens, failed...)
	}
	return failedTokens, nil
}

// SendToTopic sends a notification to all devices subscribed to an FCM topic.
// Used for blast/broadcast notifications where devices self-subscribe to topics.
//
// topic: FCM topic name without the "/topics/" prefix (e.g., "announcements")
func (c *Client) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	_, err := c.messaging.Send(ctx, &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	})
	return err
}

// IsUnregisteredToken reports whether an FCM error indicates a permanently invalid token.
// Callers should mark such tokens as stale in their token store.
func IsUnregisteredToken(err error) bool {
	return messaging.IsRegistrationTokenNotRegistered(err)
}

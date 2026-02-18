package notification

import (
	"context"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"

	"ichi-go/config"
	notifChannels "ichi-go/internal/applications/notification/channels"
	notifController "ichi-go/internal/applications/notification/controller"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/notification/fcm"
	notiftemplate "ichi-go/pkg/notification/template"
	// Import builtin templates so their init() functions run and register them.
	_ "ichi-go/pkg/notification/template/builtin"
)

// RegisterProviders registers all notification domain dependencies with the DI injector.
func RegisterProviders(injector do.Injector) {
	do.Provide(injector, ProvideTemplateRegistry)
	do.Provide(injector, ProvideTemplateOverrideRepository)
	do.Provide(injector, ProvideCampaignRepository)
	do.Provide(injector, ProvideLogRepository)
	do.Provide(injector, ProvideTemplateRenderer)
	do.Provide(injector, ProvideMainProducer)
	do.ProvideNamed(injector, "notification.blast.producer", ProvideBlastProducer)
	do.ProvideNamed(injector, "notification.user.producer", ProvideUserProducer)
	do.Provide(injector, ProvideFCMClient)
	do.Provide(injector, ProvidePushChannel)
	do.Provide(injector, ProvideCampaignService)
	do.Provide(injector, ProvideNotificationService)
	do.Provide(injector, ProvideNotificationController)
}

// ProvideTemplateRegistry returns the global Go template registry.
// Builtin templates are registered via init() when the builtin package is imported above.
func ProvideTemplateRegistry(_ do.Injector) (*notiftemplate.Registry, error) {
	return notiftemplate.GlobalRegistry, nil
}

func ProvideTemplateOverrideRepository(i do.Injector) (*repositories.NotificationTemplateOverrideRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewNotificationTemplateOverrideRepository(db), nil
}

func ProvideCampaignRepository(i do.Injector) (*repositories.NotificationCampaignRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewNotificationCampaignRepository(db), nil
}

func ProvideLogRepository(i do.Injector) (*repositories.NotificationLogRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewNotificationLogRepository(db), nil
}

func ProvideTemplateRenderer(i do.Injector) (*services.TemplateRenderer, error) {
	registry := do.MustInvoke[*notiftemplate.Registry](i)
	overrideRepo := do.MustInvoke[*repositories.NotificationTemplateOverrideRepository](i)
	return services.NewTemplateRenderer(registry, overrideRepo), nil
}

// ProvideMainProducer returns a producer bound to the app.events (x-delayed-message) exchange.
// Used by CampaignService to publish with optional delay.
func ProvideMainProducer(i do.Injector) (rabbitmq.MessageProducer, error) {
	conn, err := do.Invoke[*rabbitmq.Connection](i)
	if err != nil || conn == nil {
		return nil, nil // Queue disabled — CampaignService handles nil producer gracefully
	}
	cfg := do.MustInvoke[*config.Config](i)
	return rabbitmq.NewProducer(conn, cfg.Queue().RabbitMQ)
}

// ProvideFCMClient initializes the Firebase Cloud Messaging client.
// Returns nil (not an error) when FCM is disabled in config so the app starts without credentials.
func ProvideFCMClient(_ do.Injector) (*fcm.Client, error) {
	if !viper.GetBool("notification.fcm.enabled") {
		return nil, nil
	}
	credFile := viper.GetString("notification.fcm.credentials_file")
	client, err := fcm.NewClient(context.Background(), credFile)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ProvidePushChannel provides the FCM-backed push channel.
// Falls back to no-op when FCM client is nil (disabled).
func ProvidePushChannel(i do.Injector) (*notifChannels.PushChannel, error) {
	fcmClient, _ := do.Invoke[*fcm.Client](i)
	return notifChannels.NewPushChannel(fcmClient), nil
}

func ProvideCampaignService(i do.Injector) (*services.CampaignService, error) {
	registry := do.MustInvoke[*notiftemplate.Registry](i)
	campaignRepo := do.MustInvoke[*repositories.NotificationCampaignRepository](i)
	producer, _ := do.Invoke[rabbitmq.MessageProducer](i)
	return services.NewCampaignService(registry, campaignRepo, producer), nil
}

func ProvideNotificationController(i do.Injector) (*notifController.NotificationController, error) {
	campaignSvc := do.MustInvoke[*services.CampaignService](i)
	return notifController.NewNotificationController(campaignSvc), nil
}

// ProvideBlastProducer returns a producer bound to the fanout blast exchange.
// Returns nil (not an error) when the queue connection is unavailable.
func ProvideBlastProducer(i do.Injector) (rabbitmq.MessageProducer, error) {
	conn, err := do.Invoke[*rabbitmq.Connection](i)
	if err != nil || conn == nil {
		return nil, nil // Queue disabled — NotificationService handles nil producer gracefully
	}
	cfg := do.MustInvoke[*config.Config](i)
	rmqCfg := cfg.Queue().RabbitMQ
	rmqCfg.Publisher.ExchangeName = "notification.blast"
	return rabbitmq.NewProducer(conn, rmqCfg)
}

// ProvideUserProducer returns a producer bound to the topic user exchange.
// Returns nil (not an error) when the queue connection is unavailable.
func ProvideUserProducer(i do.Injector) (rabbitmq.MessageProducer, error) {
	conn, err := do.Invoke[*rabbitmq.Connection](i)
	if err != nil || conn == nil {
		return nil, nil // Queue disabled — NotificationService handles nil producer gracefully
	}
	cfg := do.MustInvoke[*config.Config](i)
	rmqCfg := cfg.Queue().RabbitMQ
	rmqCfg.Publisher.ExchangeName = "notification.user"
	return rabbitmq.NewProducer(conn, rmqCfg)
}

// ProvideNotificationService wires NotificationService with its blast and user producers.
func ProvideNotificationService(i do.Injector) (*services.NotificationService, error) {
	blastProducer, _ := do.InvokeNamed[rabbitmq.MessageProducer](i, "notification.blast.producer")
	userProducer, _ := do.InvokeNamed[rabbitmq.MessageProducer](i, "notification.user.producer")
	return services.NewNotificationService(blastProducer, userProducer), nil
}

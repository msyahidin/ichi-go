package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/models"
	"ichi-go/internal/infra/queue/rabbitmq"
	notiftemplate "ichi-go/pkg/notification/template"
)

// ============================================================================
// Mocks
// ============================================================================

// mockCampaignRepo satisfies the CampaignRepository interface.
type mockCampaignRepo struct {
	mock.Mock
}

func (m *mockCampaignRepo) CreateCampaign(ctx context.Context, campaign *models.NotificationCampaign) (*models.NotificationCampaign, error) {
	args := m.Called(ctx, campaign)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationCampaign), args.Error(1)
}

func (m *mockCampaignRepo) UpdateStatus(ctx context.Context, id int64, status models.CampaignStatus, errMsg string, publishedAt *time.Time) error {
	args := m.Called(ctx, id, status, errMsg, publishedAt)
	return args.Error(0)
}

// mockEventTemplate implements notiftemplate.EventTemplate.
type mockEventTemplate struct {
	slug     string
	channels []string
}

func (m *mockEventTemplate) Slug() string             { return m.slug }
func (m *mockEventTemplate) SupportedChannels() []string { return m.channels }
func (m *mockEventTemplate) DefaultContent(channel, locale string) notiftemplate.ChannelContent {
	return notiftemplate.ChannelContent{Title: "Test Title", Body: "Test Body"}
}

// ============================================================================
// Helpers
// ============================================================================

// newTestRegistry creates a fresh isolated registry with a single test template.
func newTestRegistry(slug string, channels []string) *notiftemplate.Registry {
	reg := notiftemplate.NewRegistry()
	reg.Register(&mockEventTemplate{slug: slug, channels: channels})
	return reg
}

// createdCampaignWithID returns a campaign stub simulating the DB INSERT result.
// mode is required so publish() can switch on DeliveryMode.
// channels must match req.Channels (as []string) so publish() reads the correct channels.
func createdCampaignWithID(id int64, mode dto.DeliveryMode, channels []string) *models.NotificationCampaign {
	c := &models.NotificationCampaign{
		Status:       models.CampaignStatusPending,
		DeliveryMode: string(mode),
		Channels:     channels,
	}
	c.ID = id
	return c
}

func u32(v uint32) *uint32 { return &v }
func timePtr(t time.Time) *time.Time { return &t }

// setupCampaignSvc creates a CampaignService backed by a test registry, mock repo, and mock producer.
func setupCampaignSvc(t *testing.T, slug string) (*CampaignService, *mockCampaignRepo, *mockProducer) {
	t.Helper()
	reg := newTestRegistry(slug, []string{"email", "push"})
	repo := new(mockCampaignRepo)
	producer := new(mockProducer)
	svc := NewCampaignService(reg, repo, producer)
	return svc, repo, producer
}

func baseReq(slug string, mode dto.DeliveryMode) dto.SendNotificationRequest {
	return dto.SendNotificationRequest{
		EventSlug:    slug,
		DeliveryMode: mode,
		Channels:     []dto.Channel{dto.ChannelEmail},
	}
}

// ============================================================================
// validateChannels — pure helper
// ============================================================================

func TestValidateChannels(t *testing.T) {
	tests := []struct {
		name      string
		requested []dto.Channel
		supported []string
		wantErr   bool
		errContains string
	}{
		{
			name:      "subset of supported",
			requested: []dto.Channel{dto.ChannelEmail},
			supported: []string{"email", "push"},
			wantErr:   false,
		},
		{
			name:      "all supported",
			requested: []dto.Channel{dto.ChannelEmail, dto.ChannelPush},
			supported: []string{"email", "push"},
			wantErr:   false,
		},
		{
			name:        "unsupported channel",
			requested:   []dto.Channel{"sms"},
			supported:   []string{"email", "push"},
			wantErr:     true,
			errContains: "channels_not_supported: sms",
		},
		{
			name:        "mixed: one valid one invalid",
			requested:   []dto.Channel{dto.ChannelEmail, "sms"},
			supported:   []string{"email", "push"},
			wantErr:     true,
			errContains: "sms",
		},
		{
			name:      "empty requested",
			requested: []dto.Channel{},
			supported: []string{"email"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannels(tt.requested, tt.supported)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// resolveDelay — pure helper
// ============================================================================

func TestResolveDelay(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	wayFuture := time.Now().Add(25 * 24 * time.Hour)

	tests := []struct {
		name        string
		scheduledAt *time.Time
		delaySeconds *uint32
		wantDuration time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name:         "no scheduling",
			wantDuration: 0,
			wantErr:      false,
		},
		{
			name:        "mutually exclusive",
			scheduledAt:  timePtr(future),
			delaySeconds: u32(60),
			wantErr:     true,
			errContains: "mutually exclusive",
		},
		{
			name:         "future scheduled_at",
			scheduledAt:  timePtr(future),
			wantDuration: -1, // just assert > 0
			wantErr:      false,
		},
		{
			name:        "past scheduled_at",
			scheduledAt: timePtr(past),
			wantErr:     true,
			errContains: "must be in the future",
		},
		{
			name:        "scheduled_at too far ahead",
			scheduledAt: timePtr(wayFuture),
			wantErr:     true,
			errContains: "exceeds maximum delay",
		},
		{
			name:         "valid delay_seconds",
			delaySeconds: u32(60),
			wantDuration: 60 * time.Second,
			wantErr:      false,
		},
		{
			name:         "zero delay_seconds",
			delaySeconds: u32(0),
			wantDuration: 0,
			wantErr:      false,
		},
		{
			name:        "delay_seconds exceeds max",
			delaySeconds: u32(maxDelaySeconds + 1),
			wantErr:     true,
			errContains: "must not exceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := resolveDelay(tt.scheduledAt, tt.delaySeconds)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				if tt.wantDuration == -1 {
					assert.Greater(t, d, time.Duration(0))
				} else {
					assert.Equal(t, tt.wantDuration, d)
				}
			}
		})
	}
}

// ============================================================================
// applyExclusions — pure helper
// ============================================================================

func TestApplyExclusions(t *testing.T) {
	tests := []struct {
		name      string
		targets   []int64
		excludes  []int64
		expected  []int64
	}{
		{"no exclusions", []int64{1, 2, 3}, nil, []int64{1, 2, 3}},
		{"exclude one", []int64{1, 2, 3}, []int64{2}, []int64{1, 3}},
		{"exclude all", []int64{1, 2, 3}, []int64{1, 2, 3}, []int64{}},
		{"empty targets", []int64{}, []int64{1}, []int64{}},
		{"exclude not in targets", []int64{1, 2}, []int64{99}, []int64{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyExclusions(tt.targets, tt.excludes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Send() — validation failures (no DB/queue calls)
// ============================================================================

func TestSend_UnregisteredSlug(t *testing.T) {
	svc, repo, _ := setupCampaignSvc(t, "order.shipped")
	req := baseReq("promo.flash", dto.DeliveryModeBlast)

	_, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event_not_registered")
	repo.AssertNotCalled(t, "CreateCampaign")
}

func TestSend_UnsupportedChannel(t *testing.T) {
	svc, repo, _ := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)
	req.Channels = []dto.Channel{"sms"}

	_, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "channels_not_supported")
	repo.AssertNotCalled(t, "CreateCampaign")
}

func TestSend_MutuallyExclusiveSchedule(t *testing.T) {
	svc, repo, _ := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)
	req.ScheduledAt = timePtr(time.Now().Add(1 * time.Hour))
	req.DelaySeconds = u32(60)

	_, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	repo.AssertNotCalled(t, "CreateCampaign")
}

func TestSend_PastScheduledAt(t *testing.T) {
	svc, repo, _ := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)
	req.ScheduledAt = timePtr(time.Now().Add(-1 * time.Hour))

	_, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be in the future")
	repo.AssertNotCalled(t, "CreateCampaign")
}

func TestSend_DBCreateFails(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)

	repo.On("CreateCampaign", mock.Anything, mock.AnythingOfType("*models.NotificationCampaign")).
		Return(nil, errors.New("db timeout"))

	_, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to persist campaign")
	producer.AssertNotCalled(t, "Publish")
}

// ============================================================================
// Send() — happy paths
// ============================================================================

func TestSend_BlastHappyPath(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusPublished, "", mock.AnythingOfType("*time.Time")).Return(nil)
	producer.On("Publish", mock.Anything, dispatchRoutingKey, mock.MatchedBy(func(e dto.NotificationEvent) bool {
		return e.EventID == "campaign-7-blast" && e.DeliveryMode == dto.DeliveryModeBlast
	}), mock.Anything).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, models.CampaignStatusPublished, campaign.Status)
	assert.NotNil(t, campaign.PublishedAt)
	producer.AssertNumberOfCalls(t, "Publish", 1)
	repo.AssertExpectations(t)
}

func TestSend_UserHappyPath(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeUser)
	req.UserTargetIDs = []int64{1, 2, 3}

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusPublished, "", mock.AnythingOfType("*time.Time")).Return(nil)
	producer.On("Publish", mock.Anything, dispatchRoutingKey, mock.MatchedBy(func(e dto.NotificationEvent) bool {
		return e.DeliveryMode == dto.DeliveryModeUser
	}), mock.Anything).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, models.CampaignStatusPublished, campaign.Status)
	producer.AssertNumberOfCalls(t, "Publish", 3) // one per user
}

func TestSend_UserWithExclusion(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeUser)
	req.UserTargetIDs = []int64{1, 2, 3}
	req.UserExcludeIDs = []int64{2}

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	producer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	producer.AssertNumberOfCalls(t, "Publish", 2) // user 1 and 3, not 2
}

func TestSend_AllUsersExcluded(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeUser)
	req.UserTargetIDs = []int64{1}
	req.UserExcludeIDs = []int64{1}

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusPublished, "", mock.AnythingOfType("*time.Time")).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, models.CampaignStatusPublished, campaign.Status)
	producer.AssertNotCalled(t, "Publish") // no users remain
}

// ============================================================================
// Send() — failure paths
// ============================================================================

func TestSend_NilProducer(t *testing.T) {
	reg := newTestRegistry("order.shipped", []string{"email"})
	repo := new(mockCampaignRepo)
	svc := NewCampaignService(reg, repo, nil) // nil producer

	req := baseReq("order.shipped", dto.DeliveryModeBlast)

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusFailed, mock.AnythingOfType("string"), (*time.Time)(nil)).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Equal(t, models.CampaignStatusFailed, campaign.Status)
	repo.AssertExpectations(t)
}

func TestSend_PublishFails(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusFailed, mock.AnythingOfType("string"), (*time.Time)(nil)).Return(nil)
	producer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("broker down"))

	campaign, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Equal(t, models.CampaignStatusFailed, campaign.Status)
	assert.NotEmpty(t, campaign.ErrorMessage)
	repo.AssertExpectations(t)
}

func TestSend_UpdateStatusFails(t *testing.T) {
	// When UpdateStatus fails after a successful publish, Send must return nil campaign + the DB error.
	// The in-memory campaign must NOT be mutated to published state.
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusPublished, "", mock.AnythingOfType("*time.Time")).
		Return(errors.New("db connection lost"))
	producer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, campaign) // no partially-mutated campaign returned to caller
	repo.AssertExpectations(t)
}

func TestSend_DelaySeconds(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)
	req.DelaySeconds = u32(30)

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	producer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		return opts.Delay == 30*time.Second
	})).Return(nil)

	_, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	producer.AssertExpectations(t)
}

func TestSend_DefaultLocale(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", dto.DeliveryModeBlast)
	req.Locale = "" // empty — should default to "en"

	var capturedCampaign *models.NotificationCampaign
	repo.On("CreateCampaign", mock.Anything, mock.MatchedBy(func(c *models.NotificationCampaign) bool {
		capturedCampaign = c
		return true
	})).Return(createdCampaignWithID(7, req.DeliveryMode, []string{"email"}), nil)
	repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	producer.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Send(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, capturedCampaign)
	assert.Equal(t, "en", capturedCampaign.Locale)
}

func TestSend_UnknownDeliveryMode(t *testing.T) {
	svc, repo, producer := setupCampaignSvc(t, "order.shipped")
	req := baseReq("order.shipped", "webhook") // invalid mode

	returned := createdCampaignWithID(7, req.DeliveryMode, []string{"email"})
	repo.On("CreateCampaign", mock.Anything, mock.Anything).Return(returned, nil)
	repo.On("UpdateStatus", mock.Anything, int64(7), models.CampaignStatusFailed, mock.AnythingOfType("string"), (*time.Time)(nil)).Return(nil)

	campaign, err := svc.Send(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown delivery_mode")
	assert.Equal(t, models.CampaignStatusFailed, campaign.Status)
	producer.AssertNotCalled(t, "Publish")
}

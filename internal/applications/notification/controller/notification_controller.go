package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/services"
	"ichi-go/pkg/utils/response"
)

// NotificationController handles the notification send API.
type NotificationController struct {
	campaignService *services.CampaignService
}

func NewNotificationController(campaignService *services.CampaignService) *NotificationController {
	return &NotificationController{campaignService: campaignService}
}

// Send handles POST /api/notifications/send.
//
// Validates the request, persists a campaign record, and publishes the notification
// event(s) to RabbitMQ with optional schedule/delay.
//
// Returns 201 Created on success with campaign_id, status, and published_at.
// Returns 422 Unprocessable Entity when event_slug is not registered or channels are unsupported.
// Returns 400 Bad Request for validation failures (missing fields, bad delay, etc.).
func (c *NotificationController) Send(eCtx echo.Context) error {
	var req dto.SendNotificationRequest
	if err := eCtx.Bind(&req); err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	// Validate required fields manually (supplement to struct tags).
	if err := validateSendRequest(&req); err != nil {
		return response.Error(eCtx, http.StatusBadRequest, err)
	}

	campaign, err := c.campaignService.Send(eCtx.Request().Context(), req)
	if err != nil {
		// Distinguish between "event not registered" (422) and other errors (500).
		if isEventNotRegisteredError(err) {
			return response.Error(eCtx, http.StatusUnprocessableEntity, err)
		}
		if isValidationError(err) {
			return response.Error(eCtx, http.StatusBadRequest, err)
		}
		return response.Error(eCtx, http.StatusInternalServerError, err)
	}

	return response.Created(eCtx, dto.SendNotificationResponse{
		CampaignID:  campaign.ID,
		Status:      string(campaign.Status),
		PublishedAt: campaign.PublishedAt,
	})
}

// validateSendRequest performs cross-field validation not expressible in struct tags.
func validateSendRequest(req *dto.SendNotificationRequest) error {
	// user_target_ids required when delivery_mode=user.
	if req.DeliveryMode == dto.DeliveryModeUser && len(req.UserTargetIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "user_target_ids is required when delivery_mode=user")
	}

	return nil
}

// isEventNotRegisteredError detects the "not registered in template registry" error.
func isEventNotRegisteredError(err error) bool {
	return strings.Contains(err.Error(), "event_not_registered") ||
		strings.Contains(err.Error(), "is not registered")
}

// isValidationError detects service-layer validation errors (delay, channels, etc.).
func isValidationError(err error) bool {
	s := err.Error()
	return strings.HasPrefix(s, "channels_not_supported:") ||
		strings.HasPrefix(s, "scheduled_at") ||
		strings.HasPrefix(s, "delay_seconds") ||
		strings.Contains(s, "mutually exclusive") ||
		strings.Contains(s, "must be in the future")
}

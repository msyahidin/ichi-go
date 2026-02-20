package repositories_test

import (
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
)

// Compile-time assertion: *NotificationCampaignRepository must satisfy services.CampaignRepository.
// If the interface or the repository diverge in signature, this file will fail to compile.
var _ services.CampaignRepository = (*repositories.NotificationCampaignRepository)(nil)

package user

type WelcomeNotificationMessage struct {
	EventType string `json:"event_type"` // Always "user.welcome"
	UserID    string `json:"user_id"`    // User ID as string
	Email     string `json:"email"`      // User email
	Text      string `json:"text"`       // Welcome message text
}

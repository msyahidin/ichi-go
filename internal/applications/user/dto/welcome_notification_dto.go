package user

type WelcomeNotificationMessage struct {
	EventType string `json:"event_type"`
	UserId    string `json:"user_id"`
	Email     string `json:"email"`
	Text      string `json:"text"`
}

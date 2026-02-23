package entity

type WAWebhookEvent struct {
	Event    string    `json:"event"`
	DeviceID string    `json:"device_id"`
	Payload  WAMessage `json:"payload"`
}

type WAMessage struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	From      string `json:"from"`
	FromName  string `json:"from_name"`
	IsFromMe  bool   `json:"is_from_me"`
	Body      string `json:"body"`
	Audio     string `json:"audio"`
	Image     string `json:"image"`
	Timestamp string `json:"timestamp"`
}

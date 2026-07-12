package dto

type Channel struct {
	ID       string `json:"id"`
	Label    string `json:"label" binding:"required"`
	BotToken string `json:"bot_token" binding:"required"`
	ChatID   string `json:"chat_id" binding:"required"`
}

type ChannelsResponse struct {
	Channels []Channel `json:"channels"`
}

type SaveChannelsRequest struct {
	Channels []Channel `json:"channels"`
}

type RoutesResponse struct {
	Routes map[string]string `json:"routes"`
}

type SaveRoutesRequest struct {
	Routes map[string]string `json:"routes"`
}

type TestSendRequest struct {
	ChannelID string `json:"channel_id" binding:"required"`
}

type TriggerRequest struct {
	Topic string `json:"topic" binding:"required"`
}

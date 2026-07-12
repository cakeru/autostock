package models

// Channel is one configured Telegram destination: a bot token paired with a
// chat ID. Telegram doesn't distinguish groups from personal chats at the API
// level — chat_id is chat_id either way — so a channel can point at either.
type Channel struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

// Config is one branch's full Telegram setup: its channels, and which
// channel (if any) each topic is routed to.
type Config struct {
	Channels []Channel
	Routes   map[string]string // topic -> channel id, "" or absent = off
}

func (c *Config) ChannelByID(id string) (Channel, bool) {
	for _, ch := range c.Channels {
		if ch.ID == id {
			return ch, true
		}
	}
	return Channel{}, false
}

// RouteChannel resolves a topic straight to its channel, or ok=false if the
// topic isn't routed anywhere (disabled) or points at a channel that no
// longer exists.
func (c *Config) RouteChannel(topic string) (Channel, bool) {
	id := c.Routes[topic]
	if id == "" {
		return Channel{}, false
	}
	return c.ChannelByID(id)
}

// Topics this build knows how to route. Jobs/Sales/Alerts are event-driven
// (logged by the business services as things happen); the rest are
// time-based, produced by the scheduler.
const (
	TopicJobs          = "jobs"
	TopicSales         = "sales"
	TopicAlerts        = "alerts"
	TopicDailyDigest   = "daily_digest"
	TopicTomorrowAppts = "tomorrow_appts"
	TopicWeeklyAP      = "weekly_ap"
	TopicMonthlyReport = "monthly_report"
	TopicMonthlyBackup = "monthly_backup"
	TopicDueForService = "due_for_service"
	// Documents an admin sends on demand (an invoice or a vehicle report PDF) to
	// forward on to a customer — not event- or schedule-driven.
	TopicDocuments = "documents"
)

var AllTopics = []string{
	TopicJobs, TopicSales, TopicAlerts,
	TopicDailyDigest, TopicTomorrowAppts, TopicWeeklyAP,
	TopicMonthlyReport, TopicMonthlyBackup, TopicDueForService,
	TopicDocuments,
}

// ScheduledTopics are the ones the scheduler (not a business event) produces
// — these are what a manual "trigger now" is offered for.
var ScheduledTopics = []string{
	TopicDailyDigest, TopicTomorrowAppts, TopicWeeklyAP,
	TopicMonthlyReport, TopicMonthlyBackup, TopicDueForService,
}

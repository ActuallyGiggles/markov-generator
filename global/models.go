package global

type DiscordChannelInfo struct {
	ChannelName string
	ChannelID   string
}

type Directive struct {
	Platform         string            `json:"platform"`
	ChannelName      string            `json:"channel_name"`
	ChannelID        string            `json:"channel_id"`
	DiscordChannelID string            `json:"discord_channel_id"`
	Settings         DirectiveSettings `json:"settings"`
}

type DirectiveSettings struct {
	Connected          bool `json:"connected"`
	IsOnlineEnabled    bool `json:"is_online_enabled"`
	IsOfflineEnabled   bool `json:"is_offline_enabled"`
	AreCommandsEnabled bool `json:"are_commands_enabled"`
	IsOptedIn          bool `json:"is_opted_in"`
}

type Resource struct {
	DiscordChannelName string `json:"discord_channel_name"`
	DiscordChannelID   string `json:"discord_channel_id"`
	DisplayMessageID   string `json:"display_message_id"`
	Content            string `json:"display_message_content"`
}

type Emote struct {
	Name string
	Url  string
}

type ThirdPartyEmotes struct {
	Name   string
	Emotes []Emote
}

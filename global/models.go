package global

type DiscordChannelInfo struct {
	ChannelName string
	ChannelID   string
}

type Directive struct {
	Platform         string
	ChannelName      string
	ChannelID        string
	DiscordChannelID string
	Settings         DirectiveSettings
}

type DirectiveSettings struct {
	Connected           bool
	IsOnlineEnabled     bool
	IsOfflineEnabled    bool
	AreCommandsEnabled  bool
	WhichChannelsToUse  string
	CustomChannelsToUse []string
}

type Resource struct {
	DiscordChannelName string
	DiscordChannelID   string
	DisplayMessageID   string
	Content            string
}

type Emote struct {
	Name string
	Url  string
}

type ThirdPartyEmotes struct {
	Name   string
	Emotes []Emote
}

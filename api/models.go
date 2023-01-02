package api

import (
	"markov-generator/platform/twitch"
)

type APIResponse struct {
	ModeUsed       string `json:"mode_used"`
	ChannelUsed    string `json:"channel_used"`
	MessageUsed    string `json:"message_used"`
	MarkovSentence string `json:"markov_sentence"`
	Error          string `json:"error"`
}

type DataSend struct {
	ChannelsUsed []twitch.Data
	ChannelsLive []ChannelsLive
}

type ChannelsLive struct {
	Name string
	Live bool
}

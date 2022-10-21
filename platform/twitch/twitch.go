package twitch

import (
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/stats"

	"github.com/gempir/go-twitch-irc/v3"
)

var client *twitch.Client

// Start creates a twitch client and connects it.
func Start(in chan platform.Message) {
	// Make unexported client use the address for the initialized client
	client = &twitch.Client{}
	client = twitch.NewClient(global.BotName, global.TwitchBotOauth)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		m := platform.Message{
			Platform:    "twitch",
			ChannelName: message.Channel,
			ChannelID:   message.ID,
			AuthorName:  message.User.Name,
			AuthorID:    message.User.ID,
			Content:     message.Message,
		}

		in <- m
	})

	for _, directive := range global.Directives {
		client.Join(directive.ChannelName)
		//stats.Log("Joined", directive.ChannelName)
	}

	err := client.Connect()
	if err != nil {
		panic(err)
	}

	stats.Log("Twitch Started")
}

// Say sends a message to a specific twitch chatroom.
func Say(channel string, message string) {
	client.Say(channel, message)
}

// Join joins a twitch chatroom.
func Join(channel string) {
	client.Join(channel)
}

// Depart departs a twitch chatroom.
func Depart(channel string) {
	client.Depart(channel)
}

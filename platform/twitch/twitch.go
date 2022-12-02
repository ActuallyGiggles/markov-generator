package twitch

import (
	"fmt"
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/stats"

	"github.com/gempir/go-twitch-irc/v3"
)

var client *twitch.Client

var totalM int

// Start creates a twitch client and connects it.
func Start(in chan platform.Message) {
	startedOver := 0

startOver:
	// Make unexported client use the address for the initialized client
	client = &twitch.Client{}
	client = twitch.NewClient(global.BotName, "oauth:"+global.TwitchOAuth)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		m := platform.Message{
			Platform:    "twitch",
			ChannelName: message.Channel,
			ChannelID:   message.ID,
			AuthorName:  message.User.Name,
			AuthorID:    message.User.ID,
			Content:     message.Message,
		}

		// var shortB string
		// var shortA string
		// var shortC string
		// totalM++

		// if len(m.ChannelName) > 10 {
		// 	shortB = m.ChannelName[:10] + "..."
		// } else {
		// 	shortB = m.ChannelName
		// }

		// if len(m.AuthorName) > 10 {
		// 	shortA = m.AuthorName[:10] + "..."
		// } else {
		// 	shortA = m.AuthorName
		// }

		// if len(m.Content) > 100 {
		// 	shortC = m.Content[:100] + "..."
		// } else {
		// 	shortC = m.Content
		// }

		// fmt.Printf("%8d %-13s | %-13s | %q\n", totalM, shortB, shortA, shortC)
		in <- m
	})

	for _, directive := range global.Directives {
		client.Join(directive.ChannelName)
	}

	err := client.Connect()
	if err != nil {
		startedOver++

		if startedOver < 5 {
			stats.Log(err.Error())
			goto startOver
		}

		fmt.Println("started over more than 5 times\nlast error:")
		panic(err)
	}
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

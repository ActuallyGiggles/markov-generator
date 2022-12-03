package handlers

import (
	"fmt"
	"markov-generator/global"
	"markov-generator/markov"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"strings"
)

func MsgHandler(c chan platform.Message) {
	for msg := range c {
		if msg.Platform == "twitch" {

			newMessage, passed := prepareMessage(msg)
			if passed {

				exists := false

				for _, directive := range global.Directives {
					if directive.ChannelName == msg.ChannelName {
						exists = true

						if directive.Settings.Connected {
							go markov.In(msg.ChannelName, newMessage)
							go createDefaultSentence(msg.ChannelName)
						}

						msg.Content = newMessage
						if strings.Contains(msg.Content, global.BotName) {
							go createMentioningSentence(msg, directive)
						} else {
							go createImmitationSentence(msg, directive)
						}
					}
				}

				if !exists {
					discord.Say("error-tracking", fmt.Sprintf("%s does not exist as a directive", msg.ChannelName))
				}
			}

			continue
		} else if msg.Platform == "discord" {
			go discordCommands(msg)

			continue
		}
	}
}

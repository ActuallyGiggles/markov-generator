package handlers

import (
	"markov-generator/global"
	"markov-generator/markov"
	"markov-generator/platform"
	"strings"
)

func MsgHandler(c chan platform.Message) {
	for msg := range c {
		if msg.Platform == "twitch" {
			newMessage, passed := prepareMessage(msg)
			if passed {
				go markov.In(msg.ChannelName, newMessage)
				go createDefaultSentence(msg.ChannelName)

				for _, directive := range global.Directives {
					if directive.ChannelName == msg.ChannelName {
						msg.Content = newMessage
						if strings.Contains(msg.Content, global.BotName) {
							go createMentioningSentence(msg, directive)
						} else {
							go createImmitationSentence(msg, directive)
						}
					}
				}
			}

			continue
		} else if msg.Platform == "discord" {
			go discordCommands(msg)

			continue
		}
	}
}

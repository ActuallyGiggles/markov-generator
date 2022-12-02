package handlers

import (
	"markov-generator/global"
	"markov-generator/markov"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"time"
)

func outputTicker() {
	for range time.Tick(5 * time.Minute) {
		chains := markov.CurrentChains()

		for _, chain := range chains {
			if chain == "actuallygiggles" {
				continue
			}

			createDefaultSentence(chain)

			continue
		}
	}
}

func OutputHandler(origin string, channelUsed string, sendBackToChannel string, message string, mention string) {
	str := "Channel: " + channelUsed + "\nMessage: " + message

	discord.Say("all", str)
	discord.Say(channelUsed, message)

	if global.Regex.MatchString(message) {
		discord.Say("quarantine", str)
	} else {
		twitter.AddMessageToPotentialTweets(channelUsed, message)

		if origin == "createImmitationSentence" {
			twitch.Say(sendBackToChannel, message)
		} else if origin == "createMentioningSentence" {
			twitch.Say(sendBackToChannel, "@"+mention+" "+message)
			discord.Say("mentioning", "Channel Used: "+channelUsed+"\nChannel Sent To: "+sendBackToChannel+"\nMessage: @"+mention+" "+message)
		}
	}

	if origin == "api" {
		discord.Say("website-results", str)
	}

	return
}

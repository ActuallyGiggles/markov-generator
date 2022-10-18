package handler

import (
	"log"
	"markov-generator/commands"
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"strings"
	"sync"
	"time"

	"markov-generator/markov"
)

var (
	channelLock   = make(map[string]bool)
	channelLockMx sync.Mutex
	recursions    = make(map[string]int)
	recursionsMx  sync.Mutex
	respondLock   = make(map[string]bool)
	respondLockMx sync.Mutex
)

func MsgHandler(c chan platform.Message) {
	go outputTicker()
	for msg := range c {
		if msg.Platform == "twitch" {
			newMessage, passed := prepareMessage(msg)
			if passed {
				go markov.In(msg.ChannelName, newMessage)
				go responseWarden(msg.ChannelName, msg.Content)
				go discordWarden(msg.ChannelName)
			}
			continue
		} else if msg.Platform == "discord" {
			go commands.AdminCommands(msg)
			continue
		}
	}
}

func outputTicker() {
	for range time.Tick(5 * time.Minute) {
		chains := markov.CurrentChains()
		for _, chain := range chains {
			time.Sleep(5 * time.Second)
			discordWarden(chain)
		}
	}
}

func discordWarden(channel string) {
	if !lockChannel(60, channel) {
		return
	}
	go discordGuard(channel)
	return
}

func discordGuard(channel string) {
	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	}

	output, err := markov.Out(oi)

	if err == nil {
		if !RandomlyPickLongerSentences(output) {
			recurse(channel)
		} else {
			log.Println(channel + ": " + output)
			OutputHandler("discordWarden", channel, output)
		}
	} else {
		recurse(channel)
	}
	return
}

func recurse(channel string) {
	recursionsMx.Lock()
	recursions[channel] += 1
	if recursions[channel] > 5 {
		recursions[channel] = 0
		recursionsMx.Unlock()
	} else {
		recursionsMx.Unlock()
		go discordGuard(channel)
	}
	return
}

func responseWarden(channel string, message string) {
	for _, directive := range global.Directives {
		if directive.ChannelName == channel {
			onlineEnabled := directive.Settings.IsOnlineEnabled
			offlineEnabled := directive.Settings.IsOfflineEnabled
			random := directive.Settings.IsOptedIn

			twitch.IsLiveMx.Lock()
			live := twitch.IsLive[directive.ChannelName]
			twitch.IsLiveMx.Unlock()

			if (live && onlineEnabled) || (!live && offlineEnabled) {
				if !lockResponse(global.RandomNumber(30, 180), channel) {
					return
				}

				log.Println("Response activated in", directive.ChannelName)

				chainToUse := directive.ChannelName

				if random {
					chainToUse = GetRandomChannel(directive.ChannelName)
				}

				log.Println("Random opted:", random)
				log.Println("Chain to use:", chainToUse)

				s := strings.Split(message, " ")
				t := global.PickRandomFromSlice(s)

				log.Println("Message used:", message)
				log.Println("Target chosen:", t)

				oi := markov.OutputInstructions{
					Method: "TargetedBeginning",
					Chain:  chainToUse,
					Target: t,
				}
				output, err := markov.Out(oi)

				if err == nil {
					log.Println("Response to use:", output)
					OutputHandler("responseGuard", channel, output)
				} else {
					log.Println(err)
					discord.Say("error-tracking", err.Error())
				}
				return
			}
		}
	}
}

func OutputHandler(origin string, channel string, message string) {
	if origin == "api" {
		log.Println("api output triggered", channel, message)
		defer log.Println("api output finished", channel, message)
	}

	str := "Channel: " + channel + "\nMessage: " + message
	discord.Say("all", str)
	discord.Say(channel, message)

	if global.Regex.MatchString(message) {
		discord.Say("quarantine", str)
	} else {
		twitter.AddMessageToPotentialTweets(channel, message)
		if origin == "responseGuard" {
			twitch.Say(channel, message)
		}
	}
	return
}

package handler

import (
	"fmt"
	"markov-generator/commands"
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"markov-generator/stats"
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
	//go outputTicker()
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
		fmt.Println("Ticker started")

		chains := markov.CurrentChains()
		for _, chain := range chains {
			if chain == "actuallygiggles" {
				continue
			}
			fmt.Println(chain)
			discordGuard(chain)
			continue
		}

		fmt.Println("Ticker finished")
	}
}

func discordWarden(channel string) {
	if !lockChannel(300, channel) {
		return
	}
	discordGuard(channel)
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
			OutputHandler("discordWarden", channel, output)
		}
	} else {
		if strings.Contains(err.Error(), "not found in directory") {
			return
		}
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

				chainToUse := directive.ChannelName

				if random {
					chainToUse = GetRandomChannel(directive.ChannelName)
				}

				s := strings.Split(message, " ")
				t := global.PickRandomFromSlice(s)

				oi := markov.OutputInstructions{
					Method: "TargetedBeginning",
					Chain:  chainToUse,
					Target: t,
				}
				output, err := markov.Out(oi)

				if err == nil {
					OutputHandler("responseGuard", channel, output)
				} else {
					if strings.Contains(err.Error(), "The system cannot find the file specified.") {
						return
					}
					stats.Log(err.Error())
					discord.Say("error-tracking", err.Error())
				}
				return
			}
		}
	}
}

func OutputHandler(origin string, channel string, message string) {
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

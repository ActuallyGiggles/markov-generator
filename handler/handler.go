package handler

import (
	"MarkovGenerator/commands"
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitch"
	"MarkovGenerator/platform/twitter"
	"log"
	"strings"
	"sync"
	"time"

	"MarkovGenerator/markov"
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
				go markov.Input(msg.ChannelName, newMessage)
				go warden("message", msg.ChannelName, msg.Content)
				go SendBackOutput(msg)
			}
			continue
		} else if msg.Platform == "discord" {
			go commands.AdminCommands(msg)
			continue
		} else if msg.Platform == "api" {
			go handleSuccessfulOutput(msg.ChannelName, msg.Content)
			continue
		}
	}
}

func outputTicker() {
	for range time.Tick(2 * time.Minute) {
		chains := markov.Chains()
		for _, chain := range chains {
			warden("ticker", chain, "")
		}
	}
}

func warden(origin string, channel string, message string) {
	if !lockChannel(30, channel) {
		return
	}

	c := make(chan string)
	go guard(origin, channel, message, c)
	r := <-c
	if r == "" {
		return
	} else {
		handleSuccessfulOutput(channel, r)
	}
}

func guard(origin string, channel string, message string, c chan string) {
	oi := markov.OutputInstructions{
		Chain: channel,
	}

	if origin == "message" {
		s := strings.Split(message, " ")
		m := global.PickRandomFromSlice(s)

		oi.Method = "TargetedBeginning"
		oi.Target = m
	} else if origin == "ticker" || origin == "api" {
		oi.Method = "LikelyBeginning"
	}

	output, problem := markov.Output(oi)

	if problem == "" {
		if !RandomlyPickLongerSentences(output) {
			recurse(origin, channel, message, c)
			return
		} else {
			c <- output
			close(c)
			return
		}
	} else {
		recurse(origin, channel, message, c)
		return
	}
}

func recurse(origin string, channel string, message string, c chan string) {
	recursionsMx.Lock()
	recursions[channel] += 1
	if recursions[channel] > 10 {
		recursions[channel] = 0
		recursionsMx.Unlock()
		c <- ""
		close(c)
	} else {
		recursionsMx.Unlock()
		go guard(origin, channel, message, c)
	}
}

func handleSuccessfulOutput(channel string, message string) {
	str := "Channel: " + channel + "\nMessage: " + message
	discord.Say("all", str)
	discord.Say(channel, message)

	if global.Regex.MatchString(message) {
		discord.Say("quarantine", message)
	} else {
		twitter.AddMessageToPotentialTweets(channel, message)
	}
}

func SendBackOutput(msg platform.Message) {
	for _, directive := range global.Directives {
		if directive.ChannelName == msg.ChannelName {
			onlineEnabled := directive.Settings.IsOnlineEnabled
			offlineEnabled := directive.Settings.IsOfflineEnabled
			random := directive.Settings.IsOptedIn

			twitch.IsLiveMx.Lock()
			live := twitch.IsLive[directive.ChannelName]
			twitch.IsLiveMx.Unlock()

			if (live && onlineEnabled) || (!live && offlineEnabled) {
				if !lockResponse(global.RandomNumber(30, 180), directive.ChannelName) {
					return
				}

				log.Println("Response activated in", directive.ChannelName)

				chainToUse := directive.ChannelName

				if random {
					chainToUse = GetRandomChannel(directive.ChannelName)
				}

				log.Println("Random opted:", random)
				log.Println("Chain to use:", chainToUse)

				s := strings.Split(msg.Content, " ")
				t := global.PickRandomFromSlice(s)

				log.Println("Message used:", msg.Content)
				log.Println("Target chosen:", t)

				oi := markov.OutputInstructions{
					Method: "TargetedBeginning",
					Chain:  chainToUse,
					Target: t,
				}
				output, problem := markov.Output(oi)

				if problem != "" {
					log.Println("Problem found:", problem)
					discord.Say("error-tracking", problem)
				} else {
					log.Println("Response to use:", output)
					discord.Say("all", output)
					discord.Say(chainToUse, output)

					if global.Regex.MatchString(output) {
						discord.Say("quarantine", output)
					} else {
						twitter.AddMessageToPotentialTweets(chainToUse, output)
						twitch.Say(directive.ChannelName, output)
					}
				}
				return
			}
		}
	}
}

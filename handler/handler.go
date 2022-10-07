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
				go markov.Input(msg.ChannelName, newMessage)
				go warden("message", msg.ChannelName, msg.Content)
			}
			continue
		} else if msg.Platform == "discord" {
			go commands.AdminCommands(msg)
			continue
		} else if msg.Platform == "api" {
			go outputHandler("api", msg.ChannelName, msg.Content)
			continue
		}
	}
}

func outputTicker() {
	for range time.Tick(2 * time.Minute) {
		chains := markov.CurrentChains()
		for _, chain := range chains {
			log.Println("outputTicker for", chain)
			go warden("ticker", chain, "")
		}
	}
}

func warden(origin string, channel string, message string) {
	if origin == "message" {
		if !lockResponse(global.RandomNumber(30, 180), channel) {
			return
		}

		go responseGuard(channel, message)
	} else {
		if !lockChannel(30, channel) {
			return
		}

		c := make(chan string)
		go discordGuard(channel, message, c)
		r := <-c

		if r == "" {
			return
		} else {
			outputHandler("discordGuard", channel, r)
		}
	}
}

func discordGuard(channel string, message string, c chan string) {
	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	}

	output, problem := markov.Output(oi)

	if problem == "" {
		if !RandomlyPickLongerSentences(output) {
			recurse(channel, message, c)
			return
		} else {
			c <- output
			close(c)
			return
		}
	} else {
		recurse(channel, message, c)
		return
	}
}

func recurse(channel string, message string, c chan string) {
	recursionsMx.Lock()
	recursions[channel] += 1
	if recursions[channel] > 10 {
		recursions[channel] = 0
		recursionsMx.Unlock()
		c <- ""
		close(c)
		return
	} else {
		recursionsMx.Unlock()
		go discordGuard(channel, message, c)
		return
	}
}

func responseGuard(channel string, message string) {
	for _, directive := range global.Directives {
		if directive.ChannelName == channel {
			onlineEnabled := directive.Settings.IsOnlineEnabled
			offlineEnabled := directive.Settings.IsOfflineEnabled
			random := directive.Settings.IsOptedIn

			twitch.IsLiveMx.Lock()
			live := twitch.IsLive[directive.ChannelName]
			twitch.IsLiveMx.Unlock()

			if (live && onlineEnabled) || (!live && offlineEnabled) {
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
				output, problem := markov.Output(oi)

				if problem != "" {
					log.Println("Problem found:", problem)
					discord.Say("error-tracking", problem)
				} else {
					outputHandler("responseGuard", channel, output)
					log.Println("Response to use:", output)
				}
				return
			}
		}
	}
}

func outputHandler(origin string, channel string, message string) {
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
}

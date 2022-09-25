package handler

import (
	"MarkovGenerator/commands"
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitter"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ActuallyGiggles/go-markov"
)

var (
	channelLock   = make(map[string]bool)
	channelLockMx sync.Mutex
	recursions    = make(map[string]int)
	recursionsMx  sync.Mutex
)

func MsgHandler(c chan platform.Message) {
	go outputTicker()
	for msg := range c {
		if msg.Platform == "twitch" {
			newMessage, passed := prepareMessage(msg)
			if passed {
				markov.Input(msg.ChannelName, newMessage)
				go warden("message", msg.ChannelName, msg.Content)
			}
		} else if msg.Platform == "discord" {
			commands.AdminCommands(msg)
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
	if !lockChannel(10, channel) {
		return
	}

	c := make(chan string)
	go guard(origin, channel, message, c)
	r := <-c
	if r == "" {
		return
	} else {
		str := "Channel: " + channel + "\nMessage: " + r
		discord.Say("all", str)
		discord.Say(channel, r)

		if global.Regex.MatchString(r) {
			discord.Say("quarantine", r)
		} else {
			twitter.AddMessageToPotentialTweets(channel, r)
		}
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
	} else if origin == "ticker" {
		oi.Method = "LikelyBeginning"
	}

	output, problem := markov.Output(oi)

	if problem == "" {
		if !randomlyPickLongerSentences(output) {
			recurse(origin, channel, message, c)
			return
		} else {
			fmt.Println("SUCCESS", channel, recursions[channel])
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
	fmt.Println("FAIL", channel, recursions[channel])
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

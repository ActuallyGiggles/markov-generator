package handlers

import (
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/stats"
	"strings"
	"sync"

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

// createDefaultSentence outputs a likely sentence to a channel.
func createDefaultSentence(channel string, isNewRequest bool) {
	if !lockChannel(300, channel) && isNewRequest {
		return
	}

	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	}

	output, err := markov.Out(oi)

	if err == nil {
		if RandomlyPickLongerSentences(output) {
			OutputHandler("createDefaultSentence", channel, channel, output, "")
			return
		}
	} else {
		if strings.Contains(err.Error(), "not found in directory") {
			return
		}
	}

	// Recurse the request if there is an error in the output*

	recursionsMx.Lock()
	recursions[channel] += 1

	if recursions[channel] > 5 {
		recursions[channel] = 0
		recursionsMx.Unlock()
	} else {
		recursionsMx.Unlock()
		go createDefaultSentence(channel, false)
	}

	return
}

// createImmitationSentence takes in a message and outputs a targeted sentence without mentioning a user.
func createImmitationSentence(msg platform.Message, directive global.Directive) {
	onlineEnabled := directive.Settings.IsOnlineEnabled
	offlineEnabled := directive.Settings.IsOfflineEnabled
	random := directive.Settings.IsOptedIn

	twitch.IsLiveMx.Lock()
	live := twitch.IsLive[directive.ChannelName]
	twitch.IsLiveMx.Unlock()

	if (live && onlineEnabled) || (!live && offlineEnabled) {
		if !lockResponse(global.RandomNumber(30, 180), msg.ChannelName) {
			return
		}

		chainToUse := directive.ChannelName

		if random {
			chainToUse = GetRandomChannel(directive.ChannelName)
		}

		s := strings.Split(msg.Content, " ")
		t := global.PickRandomFromSlice(s)

		oi := markov.OutputInstructions{
			Method: "TargetedBeginning",
			Chain:  chainToUse,
			Target: t,
		}
		output, err := markov.Out(oi)

		if err == nil {
			OutputHandler("createImmitationSentence", chainToUse, msg.ChannelName, output, "")
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

// createMentioningSentence takes in a message and outputs a targeted sentence that directly mentions a user.
func createMentioningSentence(msg platform.Message, directive global.Directive) {
	onlineEnabled := directive.Settings.IsOnlineEnabled
	offlineEnabled := directive.Settings.IsOfflineEnabled
	commandsEnabled := directive.Settings.AreCommandsEnabled
	random := directive.Settings.IsOptedIn

	twitch.IsLiveMx.Lock()
	live := twitch.IsLive[directive.ChannelName]
	twitch.IsLiveMx.Unlock()

	if (live && onlineEnabled) || (!live && offlineEnabled) {
		if !commandsEnabled {
			return
		}

		chainToUse := directive.ChannelName

		if random {
			chainToUse = GetRandomChannel(directive.ChannelName)
		}

		s := strings.Split(msg.Content, " ")

		for i, w := range s {
			if strings.Contains(strings.ToLower(w), global.BotName) {
				s = global.FastRemove(s, i)
				continue
			}
		}

		t := global.PickRandomFromSlice(s)

		oi := markov.OutputInstructions{
			Method: "TargetedBeginning",
			Chain:  chainToUse,
			Target: t,
		}
		output, err := markov.Out(oi)

		if err == nil {
			OutputHandler("createMentioningSentence", chainToUse, msg.ChannelName, output, msg.AuthorName)
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

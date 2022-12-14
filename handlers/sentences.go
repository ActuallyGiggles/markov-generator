package handlers

import (
	"fmt"
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
	respondLock   = make(map[string]bool)
	respondLockMx sync.Mutex
)

// createDefaultSentence outputs a likely sentence to a channel.
func createDefaultSentence(channel string) {
	if !lockChannel(300, channel) {
		return
	}

	timesRecursed := 0

recurse:
	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	}

	output, err := markov.Out(oi)

	if err == nil {
		if !IsSentenceFiltered(output) {
			OutputHandler("createDefaultSentence", channel, channel, output, "")
			return
		}
	} else {
		if strings.Contains(err.Error(), "not found in directory") ||
			strings.Contains(err.Error(), "Currently zipping") {
			return
		}
	}

	if timesRecursed < 5 {
		timesRecursed++
		goto recurse
	}

	return
}

// createImmitationSentence takes in a message and outputs a targeted sentence without mentioning a user.
func createImmitationSentence(msg platform.Message, directive global.Directive) {
	connected := directive.Settings.Connected
	onlineEnabled := directive.Settings.IsOnlineEnabled
	offlineEnabled := directive.Settings.IsOfflineEnabled
	channelsToUse := directive.Settings.WhichChannelsToUse

	twitch.IsLiveMx.Lock()
	live := twitch.IsLive[directive.ChannelName]
	twitch.IsLiveMx.Unlock()

	if (live && onlineEnabled) || (!live && offlineEnabled) {
		if randomChance := global.RandomNumber(0, 100); randomChance > 50 {
			return
		}

		if !lockResponse(global.RandomNumber(1, 10), msg.ChannelName) {
			return
		}

		chainToUse := directive.ChannelName
		recursionLimit := 50
		timesRecursed := 0

	recurse:
		switch channelsToUse {
		default:
			chainToUse = GetRandomChannel("all", directive.ChannelName)
		case "self":
			if !connected {
				chainToUse = GetRandomChannel("all", directive.ChannelName)
			} else {
				chainToUse = directive.ChannelName
			}
		case "all", "except self":
			chainToUse = GetRandomChannel(channelsToUse, directive.ChannelName)
		case "custom":
			if len(directive.Settings.CustomChannelsToUse) > 0 {
				chainToUse = global.PickRandomFromSlice(directive.Settings.CustomChannelsToUse)
			} else {
				chainToUse = GetRandomChannel("all", directive.ChannelName)
			}
		}

		method := global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
		target := removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))

		oi := markov.OutputInstructions{
			Chain:  chainToUse,
			Method: method,
			Target: target,
		}
		output, err := markov.Out(oi)

		if err == nil {
			OutputHandler("createImmitationSentence", chainToUse, msg.ChannelName, output, "")
		} else {
			// Recurse if expected error
			if strings.Contains(err.Error(), "The system cannot find the file specified.") ||
				strings.Contains(err.Error(), "does not contain parents that match") ||
				strings.Contains(err.Error(), "Currently zipping") ||
				strings.Contains(output, "@") {
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			} else {
				// Recurse if unique error
				stats.Log(err.Error())
				stats.Log("Could not create immitation sentence\n\t" + fmt.Sprintf("Trigger Message: %s", msg.Content))
				discord.Say("error-tracking", err.Error())
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			}
		}

		return
	}
}

// createMentioningSentence takes in a message and outputs a targeted sentence that directly mentions a user.
func createMentioningSentence(msg platform.Message, directive global.Directive) {
	connected := directive.Settings.Connected
	onlineEnabled := directive.Settings.IsOnlineEnabled
	offlineEnabled := directive.Settings.IsOfflineEnabled
	commandsEnabled := directive.Settings.AreCommandsEnabled
	channelsToUse := directive.Settings.WhichChannelsToUse

	twitch.IsLiveMx.Lock()
	live := twitch.IsLive[directive.ChannelName]
	twitch.IsLiveMx.Unlock()

	if (live && onlineEnabled) || (!live && offlineEnabled) {
		if !commandsEnabled {
			return
		}

		chainToUse := directive.ChannelName
		recursionLimit := len(markov.CurrentWorkers())
		timesRecursed := 0

	recurse:
		switch channelsToUse {
		default:
			chainToUse = GetRandomChannel("all", directive.ChannelName)
		case "self":
			if !connected {
				chainToUse = GetRandomChannel("all", directive.ChannelName)
			} else {
				chainToUse = directive.ChannelName
			}
		case "all", "except self":
			chainToUse = GetRandomChannel(channelsToUse, directive.ChannelName)
		case "custom":
			if len(directive.Settings.CustomChannelsToUse) > 0 {
				chainToUse = global.PickRandomFromSlice(directive.Settings.CustomChannelsToUse)
			} else {
				chainToUse = GetRandomChannel("all", directive.ChannelName)
			}
		}

		var method string
		var target string

		noMentionMsgContent := strings.Join(strings.Split(msg.Content, " ")[1:], " ")
		questionType := questionType(noMentionMsgContent)
		if questionType == "yes no question" {
			method = "TargetedBeginning"
			target = global.PickRandomFromSlice([]string{"yes", "no", "maybe", "absolutely", "absolutely", "who knows"})
		} else if questionType == "explanation question" {
			method = "TargetedBeginning"
			target = global.PickRandomFromSlice([]string{"because", "idk", "idc"})
		} else {
			method = global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
			target = removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))
			if target == "" {
				return
			}
		}

		instructions := markov.OutputInstructions{
			Method: method,
			Chain:  chainToUse,
			Target: target,
		}
		output, err := markov.Out(instructions)

		if err == nil && output != "" {
			OutputHandler("createMentioningSentence", chainToUse, msg.ChannelName, output, msg.AuthorName)
		} else {
			// Recurse if expected error
			if strings.Contains(err.Error(), "The system cannot find the file specified.") ||
				strings.Contains(err.Error(), "does not contain parents that match") ||
				strings.Contains(err.Error(), "Currently zipping") ||
				strings.Contains(output, "@") {
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			} else {
				// Recurse if unique error
				stats.Log(err.Error())
				stats.Log("Could not create mentioning sentence\n\t" + fmt.Sprintf("Trigger Message: %s", msg.Content))
				discord.Say("error-tracking", err.Error())
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			}
		}

		return
	}
}

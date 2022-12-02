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
		if strings.Contains(err.Error(), "not found in directory") {
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
		recursionLimit := 50
		timesRecursed := 0

	recurse:
		if random {
			chainToUse = GetRandomChannel(directive.ChannelName)
		}

		method := global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
		target := removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))

		oi := markov.OutputInstructions{
			Method: method,
			Chain:  chainToUse,
			Target: target,
		}
		output, err := markov.Out(oi)

		if err == nil {
			OutputHandler("createImmitationSentence", chainToUse, msg.ChannelName, output, "")
		} else {
			// Recurse if expected error
			if strings.Contains(err.Error(), "The system cannot find the file specified.") ||
				strings.Contains(err.Error(), "does not contain parents that match") {
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			} else {
				// Recurse if unique error
				stats.Log(err.Error())
				discord.Say("error-tracking", err.Error())
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			}

			stats.Log("Could not create mentioning sentence\n\t" + fmt.Sprintf("Trigger Message: %s", msg.Content))
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
		recursionLimit := 50
		timesRecursed := 0

	recurse:
		if random {
			chainToUse = GetRandomChannel(directive.ChannelName)
		}

		method := global.PickRandomFromSlice([]string{"TargetedBeginning", "TargetedMiddle", "TargetedEnding"})
		target := removeDeterminers(strings.ReplaceAll(msg.Content, ".", ""))

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
				strings.Contains(output, "@") {
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			} else {
				// Recurse if unique error
				stats.Log(err.Error())
				discord.Say("error-tracking", err.Error())
				if timesRecursed < recursionLimit {
					timesRecursed++
					goto recurse
				}
			}

			stats.Log("Could not create mentioning sentence\n\t" + fmt.Sprintf("Trigger Message: %s", msg.Content))
		}

		return
	}
}

package commands

import (
	"log"
	"markov-generator/global"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"strings"
)

func addDirective(returnChannelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(returnChannelID, messageID)

	if !checkArgFormatting(returnChannelID, args) {
		return
	}

	for _, c := range global.Directives {
		if c.ChannelName == args[1] {
			go discord.SayByIDAndDelete(returnChannelID, "Channel is already added.")
			return
		}
	}

	platform := args[0]
	channelName := args[1]
	connected, online, offline, commands, random := findSettings("add", args)

	platformChannelID, discordChannelID, success := findChannelIDs("add", platform, channelName, returnChannelID)
	if !success {
		return
	}

	channel := global.Directive{
		Platform:         platform,
		ChannelName:      channelName,
		ChannelID:        platformChannelID,
		DiscordChannelID: discordChannelID,
		Settings: global.DirectiveSettings{
			Connected:          connected,
			IsOnlineEnabled:    online,
			IsOfflineEnabled:   offline,
			AreCommandsEnabled: commands,
			IsOptedIn:          random,
		},
	}

	if !updateChannelTopic("add", channel, returnChannelID) {
		return
	}

	global.Directives = append(global.Directives, channel)

	go twitch.GetLiveStatuses()
	go twitch.GetEmoteController()
	twitch.Join(channelName)

	go discord.SayByIDAndDelete(returnChannelID, strings.Title(channelName)+" added successfully.")
}

func updateDirective(returnChannelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(returnChannelID, messageID)

	if !checkArgFormatting(returnChannelID, args) {
		return
	}

	platform := args[0]
	channelName := args[1]
	connected, online, offline, commands, random := findSettings("update", args)

	platformChannelID, discordChannelID, success := findChannelIDs("update", platform, channelName, returnChannelID)
	if !success {
		return
	}

	channel := global.Directive{
		Platform:         platform,
		ChannelName:      channelName,
		ChannelID:        platformChannelID,
		DiscordChannelID: discordChannelID,
		Settings: global.DirectiveSettings{
			Connected:          connected,
			IsOnlineEnabled:    online,
			IsOfflineEnabled:   offline,
			AreCommandsEnabled: commands,
			IsOptedIn:          random,
		},
	}

	if !updateChannelTopic("update", channel, returnChannelID) {
		return
	}

	success = removeDirective(channel.ChannelName)
	if !success {
		log.Println("failed to remove directive " + channel.ChannelName)
		discord.Say("error-tracking", "failed to remove directive "+channel.ChannelName)
	}
	global.Directives = append(global.Directives, channel)

	go discord.SayByIDAndDelete(returnChannelID, strings.Title(channelName)+" updated successfully.")
}

func connectionOfDirective(mode string, returnChannelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(returnChannelID, messageID)

	platform := args[0]
	channelName := args[1]

	existingArgs, success := findExistingSettings(channelName)
	if !success {
		log.Println("failed to find existing args for " + channelName)
		discord.Say("error-tracking", "failed to find existing args for "+channelName)
	}

	platformChannelID, discordChannelID, success := findChannelIDs("update", platform, channelName, returnChannelID)
	if !success {
		return
	}

	var connected bool

	if mode == "connect" {
		connected = true
	}

	channel := global.Directive{
		Platform:         platform,
		ChannelName:      channelName,
		ChannelID:        platformChannelID,
		DiscordChannelID: discordChannelID,
		Settings: global.DirectiveSettings{
			Connected:          connected,
			IsOnlineEnabled:    existingArgs[1],
			IsOfflineEnabled:   existingArgs[2],
			AreCommandsEnabled: existingArgs[3],
			IsOptedIn:          existingArgs[4],
		},
	}

	if !updateChannelTopic("update", channel, returnChannelID) {
		return
	}

	success = removeDirective(channel.ChannelName)
	if !success {
		log.Println("failed to remove directive " + channel.ChannelName)
		discord.Say("error-tracking", "failed to remove directive "+channel.ChannelName)
	}
	global.Directives = append(global.Directives, channel)

	go discord.SayByIDAndDelete(returnChannelID, strings.Title(channelName)+" updated successfully.")
}

func checkArgFormatting(returnChannelID string, args []string) (ok bool) {
	if len(args) < 2 {
		go discord.SayByIDAndDelete(returnChannelID, "Proper formatting ->\n"+`"platformName", "channelName", "isConnected", "isOnlineEnabled", "isOfflineEnabled", "areCommandsEnabled", "isRandomChannelOutput"`)
		return false
	}
	return true
}

func findExistingSettings(channelName string) (args []bool, success bool) {
	for i := 0; i < len(global.Directives); i++ {
		if global.Directives[i].ChannelName == channelName {
			args = append(args, global.Directives[i].Settings.Connected)
			args = append(args, global.Directives[i].Settings.IsOnlineEnabled)
			args = append(args, global.Directives[i].Settings.IsOfflineEnabled)
			args = append(args, global.Directives[i].Settings.AreCommandsEnabled)
			args = append(args, global.Directives[i].Settings.IsOptedIn)
			return args, true
		}
	}
	return args, false
}

func findSettings(mode string, args []string) (connected bool, online bool, offline bool, commands bool, random bool) {
	if mode == "add" {
		args = append(args, "true", "false", "false", "false", "false")
	} else {
		args = append(args, "false", "false", "false", "false", "false")
	}

	connected = true
	online = false
	offline = false
	commands = false
	random = false

	if args[2] == "false" {
		connected = false
	}
	if args[3] == "true" {
		online = true
	}
	if args[4] == "true" {
		offline = true
	}
	if args[5] == "true" {
		commands = true
	}
	if args[6] == "true" {
		random = true
	}

	return connected, online, offline, commands, random
}

func findChannelIDs(mode string, platform string, channelName string, returnChannelID string) (platformChannelID string, discordChannelID string, success bool) {
	if mode == "add" {
		if platform == "twitch" {
			c, ok := twitch.GetBroadcasterInfo(channelName)
			if !ok {
				go discord.SayByIDAndDelete(returnChannelID, "Is this a real twitch channel?")
				return "", "", false
			}
			platformChannelID = c.ID
		} else if platform == "youtube" {
			go discord.SayByIDAndDelete(returnChannelID, "YouTube support not yet added.")
			return
		} else {
			go discord.SayByIDAndDelete(returnChannelID, "Invalid platform.")
			return
		}

		c, ok := discord.CreateDiscordChannel(channelName)
		if !ok {
			go discord.SayByIDAndDelete(returnChannelID, "Failed to create Discord channel.")
			return "", "", false
		}
		discordChannelID = c.ID
	} else {
		for _, c := range global.Directives {
			if channelName == c.ChannelName {
				platformChannelID = c.ChannelID
				discordChannelID = c.DiscordChannelID
				break
			}
		}
	}
	return platformChannelID, discordChannelID, true
}

func removeDirective(channelName string) (success bool) {
	for i := 0; i < len(global.Directives); i++ {
		if global.Directives[i].ChannelName == channelName {
			global.Directives[i] = global.Directives[len(global.Directives)-1]
			global.Directives = global.Directives[:len(global.Directives)-1]
			return true
		}
	}
	return false
}

func updateChannelTopic(mode string, channel global.Directive, returnChannelID string) (success bool) {
	_, ok := discord.UpdateDirectiveChannel(channel)
	if !ok {
		if mode == "add" {
			discord.DeleteDiscordChannel(channel.ChannelName)
		}

		go discord.SayByIDAndDelete(returnChannelID, "Failed to "+mode+" Discord channel.")
		return false
	}
	return true
}

// UpdateResourceAndChannel updates a resource and the correlated Discord channel.
//
// mode = "add" || "remove"
func UpdateResourceAndChannel(resourceType string, mode string, returnChannelID string, messageID string, entries []string) {
	defer discord.DeleteDiscordMessage(returnChannelID, messageID)

	for i := 0; i < len(global.Resources); i++ {
		if resourceType == global.Resources[i].DiscordChannelName {
			switch mode {

			case "add":
				for _, entry := range entries {
					if global.Resources[i].Content == "" {
						global.Resources[i].Content = entry + " "
					} else {
						global.Resources[i].Content = global.Resources[i].Content + entry + " "
					}
				}
			case "remove":
				for _, entry := range entries {
					global.Resources[i].Content = strings.ReplaceAll(global.Resources[i].Content, entry+" ", "")
				}
			}

			_, ok := discord.UpdateResourceChannel(global.Resources[i])
			if !ok {
				go discord.SayByIDAndDelete(returnChannelID, "Failed to update "+resourceType+".")
			} else {
				go discord.SayByIDAndDelete(returnChannelID, "Successfully updated "+resourceType+".")
			}
		}
	}
	global.UpdateResourceLists()
}

package handlers

import (
	"markov-generator/global"
	"markov-generator/platform"
	"strings"
)

var (
	dialogueOngoing bool
	dialogueChannel chan Dialogue
)

type Dialogue struct {
	Arguments []string
	MessageID string
}

// discordCommands receives commands from an admin and returns a response.
func discordCommands(m platform.Message) {
	if m.AuthorID != global.DiscordOwnerID {
		return
	}

	message := strings.TrimPrefix(m.Content, global.Prefix)
	messageSlice := strings.Split(message, " ")
	command, args := messageSlice[0], messageSlice[1:]

	if dialogueOngoing {
		dialogueChannel <- Dialogue{Arguments: messageSlice, MessageID: m.MessageID}
		return
	}

	if !strings.HasPrefix(m.Content, global.Prefix) {
		return
	}

	switch command {
	// Directives settings
	case "addchannel":
		if len(args) == 2 {
			addDirectiveSimple(m.ChannelID, m.MessageID, args)
		} else {
			addDirectiveAdvanced(m.ChannelID, m.MessageID)
		}
	case "updatechannel":
		updateDirective(m.ChannelID, m.MessageID)

	// Resources settings
	case "seeregex":
		SendRegex(m.ChannelID, m.MessageID)
	case "addregex":
		AddRegex(m.ChannelID, m.MessageID, args)
	case "removeregex":
		RemoveRegex(m.ChannelID, m.MessageID, args)

	case "seebannedusers":
		SendBannedUsers(m.ChannelID, m.MessageID)
	case "addbanneduser":
		AddBannedUser(m.ChannelID, m.MessageID, args)
	case "removebanneduser":
		RemoveBannedUser(m.ChannelID, m.MessageID, args)
	}

	return
}

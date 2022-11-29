package handlers

import (
	"markov-generator/global"
	"markov-generator/platform"
	"strings"
)

// discordCommands receives commands from an admin and returns a response.
func discordCommands(m platform.Message) {
	if m.AuthorID != global.DiscordOwnerID || !strings.HasPrefix(m.Content, global.Prefix) || m.AuthorID == global.DiscordBotID {
		return
	}

	message := strings.TrimPrefix(m.Content, global.Prefix)
	messageSlice := strings.Split(message, " ")
	command, args := messageSlice[0], messageSlice[1:]

	switch command {

	case "addchannel", "ac":
		addDirective(m.ChannelID, m.MessageID, args)
	case "updatechannel", "uc":
		updateDirective(m.ChannelID, m.MessageID, args)
	case "connectchannel", "cc":
		connectionOfDirective("connect", m.ChannelID, m.MessageID, args)
	case "disconnectchannel", "dc":
		connectionOfDirective("disconnect", m.ChannelID, m.MessageID, args)

	case "addregex", "ar":
		UpdateResourceAndChannel("regex", "add", m.ChannelID, m.MessageID, args)
	case "removeregex", "rr":
		UpdateResourceAndChannel("regex", "remove", m.ChannelID, m.MessageID, args)
	case "addbanneduser", "abu":
		UpdateResourceAndChannel("banned-users", "add", m.ChannelID, m.MessageID, args)
	case "removebanneduser", "rbu":
		UpdateResourceAndChannel("banned-users", "remove", m.ChannelID, m.MessageID, args)
	}

	return
}

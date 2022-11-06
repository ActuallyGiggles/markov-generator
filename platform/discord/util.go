package discord

import (
	"log"
	"markov-generator/global"
	"markov-generator/stats"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// wrapInCodeBlock returns a message wrapped in three back ticks for Discord formatting.
func wrapInCodeBlock(message string) string {
	return "```" + message + "```"
}

// convertTextToDirective takes text and returns it as a Directive type.
//
// Returns the channel.
func convertTopicToDirective(topic string) (channel global.Directive, ok bool) {
	// twitch, aceu, 12341234, 43214321, true, false, false, false, false
	args := strings.Split(topic, ", ")

	if len(args) != 9 {
		return global.Directive{}, false
	}

	if len(args[1]) < 3 || len(args[1]) > 25 {
		return global.Directive{}, false
	}

	platform := args[0]
	channelName := args[1]
	channelId := args[2]
	dChannelId := args[3]
	connected := true
	onlinebool := false
	offlinebool := false
	commandsbool := false
	optedbool := false

	if args[5] == "true" {
		onlinebool = true
	}
	if args[6] == "true" {
		offlinebool = true
	}
	if args[7] == "true" {
		commandsbool = true
	}
	if args[8] == "true" {
		optedbool = true
	}

	channel = global.Directive{
		Platform:         platform,
		ChannelName:      channelName,
		ChannelID:        channelId,
		DiscordChannelID: dChannelId,
		Settings: global.DirectiveSettings{
			Connected:          connected,
			IsOnlineEnabled:    onlinebool,
			IsOfflineEnabled:   offlinebool,
			AreCommandsEnabled: commandsbool,
			IsOptedIn:          optedbool,
		},
	}

	return channel, true
}

// convertDirectiveToTopic converts a directive into a short string, where values are separated by commas.
func convertDirectiveToTopic(c global.Directive) (text string) {
	text = c.Platform + ", "
	text = text + c.ChannelName + ", "
	text = text + c.ChannelID + ", "
	text = text + c.DiscordChannelID + ", "
	text = text + strconv.FormatBool(c.Settings.Connected) + ", "
	text = text + strconv.FormatBool(c.Settings.IsOnlineEnabled) + ", "
	text = text + strconv.FormatBool(c.Settings.IsOfflineEnabled) + ", "
	text = text + strconv.FormatBool(c.Settings.AreCommandsEnabled) + ", "
	text = text + strconv.FormatBool(c.Settings.IsOptedIn)

	return text
}

// addDirective collects the topic of each channel in Discord, then adds it to global.Directives.
func getDirective(topic string) {
	c, ok := convertTopicToDirective(topic)
	if ok {
		global.Directives = append(global.Directives, c)
		return
	}
	return
}

// getResource collects the resources in each resource channel in Discord, then adds it to global.Resources.
func getResource(resourceType string, channel *discordgo.Channel) {
	dChannelName := channel.Name
	dChannelID := channel.ID
	displayMessageID := strings.TrimPrefix(channel.Topic, "resource ")
	var content string

	m, err := discord.ChannelMessage(dChannelID, displayMessageID)
	if err != nil {
		log.Fatal("Could not get message from ChannelMessage", err)
	}

	content = m.Content

	content = strings.ReplaceAll(content, "`", "")
	content = strings.ReplaceAll(content, ",\n", " ")
	content = strings.ReplaceAll(content, "This is where your "+resourceType+" will go! It uses new lines for each new entry.", "")

	Resource := global.Resource{
		DiscordChannelName: dChannelName,
		DiscordChannelID:   dChannelID,
		DisplayMessageID:   displayMessageID,
		Content:            content,
	}

	global.Resources = append(global.Resources, Resource)
	global.UpdateResourceLists()
}

// createResources creates banned users and regex channels and adds them to global.Resources.
func createResource(resourceType string) {
	c, ok := CreateDiscordChannel(resourceType)
	if !ok {
		log.Fatal("Could not create " + resourceType + " channel")
	}

	m := SayByID(c.ID, "This is where your "+resourceType+" will go! It uses new lines for each new entry.")

	Resource := global.Resource{
		DiscordChannelName: c.Name,
		DiscordChannelID:   c.ID,
		DisplayMessageID:   m.ID,
		Content:            "",
	}

	updateChannelTopic(Resource.DiscordChannelID, "resource "+m.ID)

	global.Resources = append(global.Resources, Resource)
}

func updateChannelTopic(discordChannelID string, topic string) (channel *discordgo.Channel, ok bool) {
	ch, err := discord.Channel(discordChannelID)
	if err != nil {
		stats.Log(err.Error())
		return nil, false
	}

	update := discordgo.ChannelEdit{
		Name:     ch.Name,
		Topic:    topic,
		NSFW:     &ch.NSFW,
		Position: ch.Position,
	}

	channel, err = discord.ChannelEdit(discordChannelID, &update)
	if err != nil {
		stats.Log(err.Error())
		return nil, false
	}

	return channel, true
}

package discord

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/twitter"

	//"MarkovGenerator/platform/twitter"

	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	in      chan platform.Message
	discord *discordgo.Session
)

func Start(ch chan platform.Message) {
	in = ch

	bot, err := discordgo.New("Bot " + global.DiscordToken)
	discord = bot

	if err != nil {
		panic(err)
	}

	err = discord.Open()

	if err != nil {
		panic(err)
	}

	ok := GetDirectivesAndResources(discord)
	if !ok {
		panic("Directives and Resources not initialized.")
	}

	discord.AddHandler(messageHandler)
	discord.AddHandler(reactionHandler)

	fmt.Println("Discord started")

	finishInit := platform.Message{
		Platform: "internal",
	}

	ch <- finishInit
}

// messageHandler receives messages and sends them into the in channel.
func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	incomingMessage := platform.Message{
		Platform:   "discord",
		ChannelID:  message.ChannelID,
		AuthorName: message.Author.Username,
		AuthorID:   message.Author.ID,
		MessageID:  message.ID,
		Content:    message.Content,
	}

	in <- incomingMessage
}

func reactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	fmt.Println("test")
	// If correct emoji and correct user
	if r.UserID == global.DiscordOwnerID && r.Emoji.Name == global.DiscordTweetEmote {
		fmt.Println("test2")
		manuallyTweet(r)
	}
}

// SayByID sends a message to a Discord channel via the channel ID.
//
// Returns the message id.
func SayByID(channelId string, message string) (id *discordgo.Message) {
	m, err := discord.ChannelMessageSend(channelId, wrapInCodeBlock(message))
	if err != nil {
		fmt.Println("	SayById failed \n", err)
	}
	return m
}

func Say(channel string, message string) {
	for k, v := range global.TotalChannels {
		if k == channel {
			_, err := discord.ChannelMessageSend(v, wrapInCodeBlock(message))
			if err != nil {
				fmt.Println("	SayById failed \n", err)
			}
		}
	}
	return
}

func SayByIDAndDelete(channelID string, message string) {
	m := SayByID(channelID, message)
	time.Sleep(time.Duration(5) * time.Second)
	DeleteDiscordMessage(channelID, m.ID)
}

// GetChannels returns a list of Discord channels connected to.
func GetChannels(session *discordgo.Session) (channels []*discordgo.Channel, err error) {
	s, err := session.GuildChannels(global.DiscordGuildID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateDiscordChannel creates a text channel in Discord by the passed name.
//
// Returns the channel, if the function was successful.
func CreateDiscordChannel(name string) (channel *discordgo.Channel, ok bool) {
	c, err := discord.GuildChannelCreate(global.DiscordGuildID, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		fmt.Println("	CreateDiscordChannel failed\n", err)
		return nil, false
	}
	return c, true
}

// DeleteDiscordChannel deletes any text channel in Discord by the passed name.
//
// Returns if the function was successful.
func DeleteDiscordChannel(name string) (ok bool) {
	for _, c := range global.Directives {
		if c.ChannelName == name {
			_, err := discord.ChannelDelete(c.ChannelID)
			if err != nil {
				fmt.Println("	DeleteDiscordChannel failed\n", err)
			}
		}
	}
	return true
}

func DeleteDiscordMessage(channelID string, messageID string) {
	err := discord.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		fmt.Println(err)
		fmt.Println(channelID, messageID)
	}
}

// GetDirectives collects the name and setting of each channel in Discord, then adds it to global.Directives.
//
// Returns if function is successful.
func GetDirectivesAndResources(session *discordgo.Session) (ok bool) {
	doBannedUsersExist := false
	doesRegexExist := false

	channels, err := GetChannels(session)
	if err != nil {
		panic(err)
	}

	for _, channel := range channels {
		channel = *&channel
		if _, ok := global.TotalChannels[channel.Name]; !ok {
			global.TotalChannels[channel.Name] = channel.ID
		}

		if channel.Topic == "" || channel.Topic == "non-directive" {
			continue
		}

		if strings.HasPrefix(channel.Topic, "resource") {
			if channel.Name == "banned-users" {
				doBannedUsersExist = true
				getResource("banned-users", channel)
			} else if channel.Name == "regex" {
				doesRegexExist = true
				getResource("regex", channel)
			}
			continue
		}

		getDirective(channel.Topic)
	}

	if !doBannedUsersExist {
		createResource("banned-users")
		fmt.Println("Created banned-users.")
	}
	if !doesRegexExist {
		createResource("regex")
		fmt.Println("Created regex.")
	}

	return true
}

// UpdateDirectiveChannel updates a directive channel topic on Discord.
//
// Returns if function was successful.
func UpdateDirectiveChannel(c global.Directive) (channel *discordgo.Channel, ok bool) {
	topic := convertDirectiveToTopic(c)

	channel, ok = updateChannelTopic(c.DiscordChannelID, topic)

	return channel, true
}

// UpdateResourceChannel updates a resource channel message on Discord.
//
// Returns if function was successful.
func UpdateResourceChannel(c global.Resource) (channel *discordgo.Message, ok bool) {
	content := "```" + strings.ReplaceAll(c.Content, " ", ",\n") + "```"

	message, err := discord.ChannelMessageEdit(c.DiscordChannelID, c.DisplayMessageID, content)
	if err != nil {
		panic(fmt.Sprintf("Could not edit %s message. %e", c.DiscordChannelName, err))
	}

	return message, true
}

func manuallyTweet(r *discordgo.MessageReactionAdd) {
	var channel string
	var message string

	// If message was sent by bot
	if messageInfo, _ := discord.ChannelMessage(r.ChannelID, r.MessageID); messageInfo.Author.ID != global.DiscordBotID {
		return
	}

	// If starts with "```Channel", get channel from message
	// Else, get channel from channel name
	if messageInfo, _ := discord.ChannelMessage(r.ChannelID, r.MessageID); strings.HasPrefix(messageInfo.Content, "```Channel") {
		s := strings.Split(strings.ReplaceAll(messageInfo.Content, "\n", " "), " ")
		channel = strings.ReplaceAll(s[1], "\n", "")
		message = strings.Join(s[3:], " ")
		message = strings.ReplaceAll(message, "```", "")
		twitter.SendTweet(channel, message)
	} else {
		c, _ := discord.Channel(r.ChannelID)
		channel = c.Name
		message = strings.ReplaceAll(messageInfo.Content, "`", "")
		twitter.SendTweet(channel, message)
	}
}

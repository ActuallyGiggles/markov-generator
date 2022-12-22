package handlers

import (
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"regexp"
	"strconv"
	"strings"
	"time"

	"markov-generator/markov"
)

// prepareMessage prepares the message to be inputted into a Markov chain
func prepareMessage(msg platform.Message) (processed string, passed bool) {
	processed = lowercaseIfNotEmote(msg.ChannelName, msg.Content)
	processed = removeWeirdTwitchCharactersAndTrim(processed)
	if !checkForMessageQuality(msg.AuthorName, processed) {
		return "", false
	}
	return processed, true
}

// lowercaseIfNotEmote takes channel and string and returns the string with everything lowercase except any emotes from that channel.
func lowercaseIfNotEmote(channel string, message string) string {
	global.EmotesMx.Lock()
	defer global.EmotesMx.Unlock()
	var new []string
	slice := strings.Split(message, " ")
	for _, word := range slice {
		match := false
		for _, emote := range global.GlobalEmotes {
			if word == emote.Name {
				match = true
				new = append(new, word)
				break
			}
		}

		if !match {
			for _, emote := range global.TwitchChannelEmotes {
				if word == emote.Name {
					match = true
					new = append(new, word)
					break
				}
			}
		}

		if !match {
			for _, c := range global.ThirdPartyChannelEmotes {
				if c.Name == channel {
					for _, emote := range c.Emotes {
						if word == emote.Name {
							match = true
							new = append(new, word)
							break
						}
					}
				}
			}
		}

		if !match {
			new = append(new, strings.ToLower(word))
		}
	}
	newMessage := strings.Join(new, " ")
	return newMessage
}

// removeWeirdTwitchCharactersAndTrim removes whitespaces that Twitch adds, such as  and 󠀀.
func removeWeirdTwitchCharactersAndTrim(message string) string {
	message = strings.ReplaceAll(message, "", "")
	message = strings.ReplaceAll(message, "󠀀", "")
	slice := strings.Fields(message)
	message = strings.Join(slice, " ")
	return message
}

// checkForUrl returns if a string contains a link/url.
func checkForUrl(urlOrNot string) bool {
	r, _ := regexp.Compile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	return r.MatchString(urlOrNot)
}

// checkForBotUser returns if a username belongs to a bot account.
func checkForBotUser(username string) bool {
	if strings.Contains(username, "bot") {
		return true
	}
	for _, v := range global.BannedUsers {
		if strings.Contains(username, v) {
			return true
		}
	}
	return false
}

// checkForCommand returns if a string is a potential command.
func checkForCommand(message string) bool {
	s := []string{"!", "%", "?", "-", ".", ",", "#", "+", "$"}
	for _, prefix := range s {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

// checkForRepitition returns if a string repeats words 3 or more times.
func checkForRepitition(message string) bool {
	wordList := strings.Fields(message)
	counts := make(map[string]int)
	for _, word := range wordList {
		_, ok := counts[word]
		if ok {
			counts[word] += 1
		} else {
			counts[word] = 1
		}
	}
	for _, number := range counts {
		if number > 2 {
			return true
		}
	}
	return false
}

// checkForMessageQuality checks if a username or message passes the vibe check.
func checkForMessageQuality(username string, message string) bool {
	// Check for url
	if checkForUrl(message) {
		return false
	}

	// Check usernames for bots
	if checkForBotUser(username) {
		return false
	}

	// Check for command
	if checkForCommand(message) {
		return false
	}

	// Check if message has too much repitition
	if checkForRepitition(message) {
		return false
	}

	return true
}

func lockChannel(timer int, channel string) bool {
	channelLockMx.Lock()
	if channelLock[channel] {
		channelLockMx.Unlock()
		return false
	}
	channelLock[channel] = true
	channelLockMx.Unlock()
	go unlockChannel(timer, channel)
	return true
}

func unlockChannel(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Second)
	channelLockMx.Lock()
	channelLock[channel] = false
	channelLockMx.Unlock()
}

func IsSentenceFiltered(sentence string) bool {
	// Split sentence into words
	s := strings.Split(sentence, " ")

	// If there are one to three words, 50% chance to pass
	if 0 < len(s) && len(s) < 4 {
		n := global.RandomNumber(0, 100)
		if n <= 50 {
			return true
		}
	}

	return false
}

func DoesSliceContainIndex(slice []string, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

// lockResponse will mark a channel as locked until the time (in minutes) has passed.
func lockResponse(time int, channel string) bool {
	respondLockMx.Lock()
	if respondLock[channel] {
		respondLockMx.Unlock()
		return false
	}
	respondLock[channel] = true
	respondLockMx.Unlock()
	go unlockResponse(time, channel)
	return true
}

func unlockResponse(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Minute)
	respondLockMx.Lock()
	respondLock[channel] = false
	respondLockMx.Unlock()
}

func GetRandomChannel(mode string, channel string) (randomChannel string) {
	var s []string

	chains := markov.CurrentChains()
	for _, chain := range chains {
		if mode == "except self" && chain == channel {
			continue
		}
		s = append(s, chain)
	}

	return global.PickRandomFromSlice(s)
}

func addDirectiveSimple(channelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	var messagesToDelete []string
	defer func() {
		for _, message := range messagesToDelete {
			discord.DeleteDiscordMessage(channelID, message)
		}
	}()

	for _, c := range global.Directives {
		if c.ChannelName == args[1] {
			discord.SayByIDAndDelete(channelID, "Channel is already added.")
			return
		}
	}

	platform := args[0]
	channelName := args[1]

	platformChannelID, discordChannelID, success := findChannelIDs("add", platform, channelName, channelID)
	if !success {
		return
	}

	channel := global.Directive{
		Platform:         platform,
		ChannelName:      channelName,
		ChannelID:        platformChannelID,
		DiscordChannelID: discordChannelID,
		Settings: global.DirectiveSettings{
			Connected:          true,
			WhichChannelsToUse: "self",
		},
	}

	go twitch.GetLiveStatuses(false)
	m := discord.SayByID(channelID, "Gathering emotes for "+channelName).ID
	messagesToDelete = append(messagesToDelete, m)

	ok := twitch.GetEmoteController(false, channel)
	if !ok {
		discord.DeleteDiscordChannel(channelName)
		discord.SayByIDAndDelete(channelID, "Could not retrieve "+channelName+"'s emotes...")
	}

	err := global.UpdateChannels("add", channel)
	if err == nil {
		twitch.Join(channelName)
		discord.SayByID(channelID, strings.Title(channelName)+" added successfully.")
	} else {
		discord.DeleteDiscordChannel(channelName)
		discord.SayByIDAndDelete(channelID, err.Error())
	}
}

func addDirectiveAdvanced(channelID string, messageID string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	var messagesToDelete []string
	defer func() {
		for _, message := range messagesToDelete {
			discord.DeleteDiscordMessage(channelID, message)
		}
	}()

	dialogueOngoing = true
	dialogueChannel = make(chan Dialogue)

	defer func() {
		dialogueOngoing = false
		dialogueChannel = nil
	}()

	channel := global.Directive{}

	// Get platform
	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "What is the platform?\n(1) Twitch\n(2) Youtube").ID)
	platform := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, platform.MessageID)
	switch platform.Arguments[0] {
	case "cancel":
		dialogueOngoing = false
		dialogueChannel = nil
		return
	case "1":
		channel.Platform = "twitch"
	case "2":
		channel.Platform = "youtube"
	}

	// Get channel name
	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "What is the channel name?").ID)
	channelName := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, channelName.MessageID)
	switch channelName.Arguments[0] {
	case "cancel":
		dialogueOngoing = false
		dialogueChannel = nil
		return
	default:
		channel.ChannelName = channelName.Arguments[0]
	}

	// Return if channel is already added
	for _, c := range global.Directives {
		if c.ChannelName == channelName.Arguments[0] {
			go discord.SayByIDAndDelete(channelID, "Channel is already added.")
			return
		}
	}

	// Get platform channel ID and discord channel ID
	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "Gathering IDs...").ID)
	platformChannelID, discordChannelID, success := findChannelIDs("add", channel.Platform, channelName.Arguments[0], channelID)
	if !success {
		return
	}
	channel.ChannelID = platformChannelID
	channel.DiscordChannelID = discordChannelID

	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "For the following options, type 0 if false and 1 if true:\n\n1. Will be collecting messages into Markov chain?\n2. Will message into online chat?\n3. Will message into offline chat?\n4. Will respond to mentions?").ID)
	boolSettings := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, boolSettings.MessageID)
	if boolSettings.Arguments[0] == "cancel" {
		dialogueOngoing = false
		dialogueChannel = nil
		return
	}
	for i, setting := range boolSettings.Arguments {
		if result, err := strconv.ParseBool(setting); err != nil {
			discord.SayByIDAndDelete(channelID, "Unable to parse settings.")
			discord.DeleteDiscordChannel(channelName.Arguments[0])
			return
		} else {
			switch i {
			case 0:
				channel.Settings.Connected = result
			case 1:
				channel.Settings.IsOnlineEnabled = result
			case 2:
				channel.Settings.IsOfflineEnabled = result
			case 3:
				channel.Settings.AreCommandsEnabled = result
			}
		}
	}

	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "What chains will this channel use to respond with?\n\nAll (1)     All except self (2)     Self (3)     Custom (4)\n\nIf custom, what are the custom channels to use?").ID)
	responseSettings := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, responseSettings.MessageID)
	if responseSettings.Arguments[0] == "cancel" {
		dialogueOngoing = false
		dialogueChannel = nil
		return
	}
	mode := responseSettings.Arguments[0]
	customChannels := responseSettings.Arguments[1:]
	switch mode {
	case "1", "all", "All":
		channel.Settings.WhichChannelsToUse = "all"
	case "2", "all except self", "All except self":
		channel.Settings.WhichChannelsToUse = "except self"
	case "3", "self", "Self":
		channel.Settings.WhichChannelsToUse = "self"
	case "4", "custom", "Custom":
		channel.Settings.WhichChannelsToUse = "custom"
		channel.Settings.CustomChannelsToUse = customChannels
	}

	go twitch.GetLiveStatuses(false)
	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "Gathering emotes for "+channelName.Arguments[0]).ID)

	ok := twitch.GetEmoteController(false, channel)
	if !ok {
		discord.DeleteDiscordChannel(channelName.Arguments[0])
		discord.SayByIDAndDelete(channelID, "Could not retrieve "+channelName.Arguments[0]+"'s emotes...")
	}

	err := global.UpdateChannels("add", channel)
	if err == nil {
		twitch.Join(channelName.Arguments[0])
		discord.SayByID(channelID, channelName.Arguments[0]+" added successfully.")
	} else {
		discord.DeleteDiscordChannel(channelName.Arguments[0])
		discord.SayByIDAndDelete(channelID, err.Error())
	}
}

func updateDirective(channelID string, messageID string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	var messagesToDelete []string
	defer func() {
		for _, message := range messagesToDelete {
			discord.DeleteDiscordMessage(channelID, message)
		}
	}()

	defer func() {
		dialogueOngoing = false
		dialogueChannel = nil
	}()

	dialogueOngoing = true
	dialogueChannel = make(chan Dialogue)

	var channel *global.Directive

	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "Which channel will you update?").ID)
	channelName := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, channelName.MessageID)
	if channelName.Arguments[0] == "cancel" {
		dialogueOngoing = false
		dialogueChannel = nil
		return
	}

	for _, directive := range *&global.Directives {
		if channelName.Arguments[0] == directive.ChannelName {
			channel = &directive
		}
	}

	messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "Which do you want to update?\n\n(1) Collecting messages for Markov chains\n(2) Allowing messages online\n(3) Allowing messages offline\n(4) Alowing to reply when mentioned\n(5) Which channels to use").ID)
	settingsToUpdate := <-dialogueChannel
	messagesToDelete = append(messagesToDelete, settingsToUpdate.MessageID)
	if settingsToUpdate.Arguments[0] == "cancel" {
		dialogueOngoing = false
		dialogueChannel = nil
		return
	}
	for _, setting := range settingsToUpdate.Arguments {
		if setting == "1" {
			channel.Settings.Connected = !channel.Settings.Connected
		}
		if setting == "2" {
			channel.Settings.IsOnlineEnabled = !channel.Settings.IsOnlineEnabled
		}
		if setting == "3" {
			channel.Settings.IsOfflineEnabled = !channel.Settings.IsOfflineEnabled
		}
		if setting == "4" {
			channel.Settings.AreCommandsEnabled = !channel.Settings.AreCommandsEnabled
		}
		if setting == "5" {
			messagesToDelete = append(messagesToDelete, discord.SayByID(channelID, "What chains will this channel use to respond with?\n\nAll (1)     All except self (2)     Self (3)     Custom (4)\n\nIf custom, what are the custom channels to use?").ID)
			responseSettings := <-dialogueChannel
			messagesToDelete = append(messagesToDelete, responseSettings.MessageID)
			if responseSettings.Arguments[0] == "cancel" {
				dialogueOngoing = false
				dialogueChannel = nil
				return
			}
			mode := responseSettings.Arguments[0]
			customChannels := responseSettings.Arguments[1:]
			switch mode {
			case "1", "all", "All":
				channel.Settings.WhichChannelsToUse = "all"
			case "2", "all except self", "All except self":
				channel.Settings.WhichChannelsToUse = "except self"
			case "3", "self", "Self":
				channel.Settings.WhichChannelsToUse = "self"
			case "4", "custom", "Custom":
				channel.Settings.WhichChannelsToUse = "custom"
				channel.Settings.CustomChannelsToUse = customChannels
			}
		}
	}

	err := global.UpdateChannels("update", *channel)
	if err == nil {
		discord.SayByID(channelID, strings.Title(channelName.Arguments[0])+" updated successfully.")
	} else {
		discord.SayByIDAndDelete(channelID, err.Error())
	}
}

func findChannelIDs(mode string, platform string, channelName string, returnChannelID string) (platformChannelID string, discordChannelID string, success bool) {
	if mode == "add" {
		if platform == "twitch" {
			c, err := twitch.GetBroadcasterInfo(channelName)
			if err != nil {
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

func AddRegex(channelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go discord.SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, regexToAdd := range args {
		exists := false
		for _, regexExisting := range global.RegexList {
			if regexExisting == regexToAdd {
				go discord.SayByIDAndDelete(channelID, regexToAdd+" is already added.")
				exists = true
			}
		}

		if !exists {
			global.RegexList = append(global.RegexList, regexToAdd)
		}
	}

	err := global.UpdateRegex()
	if err != nil {
		go discord.SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go discord.SayByIDAndDelete(channelID, "Regex successfully updated.")
	}
}

func RemoveRegex(channelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go discord.SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, regexToRemove := range args {
		exists := false
		for i, regexExisting := range global.RegexList {
			if regexToRemove == regexExisting {
				global.RegexList = global.FastRemove(global.RegexList, i)
				exists = true
				break
			}
		}

		if !exists {
			go discord.SayByIDAndDelete(channelID, regexToRemove+" is not on the list.")
		}
	}

	err := global.UpdateRegex()
	if err != nil {
		go discord.SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go discord.SayByIDAndDelete(channelID, "Regex successfully updated.")
	}
}

func SendRegex(channelID string, messageID string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)
	discord.SayByIDAndDelete(channelID, strings.Join(global.RegexList, ",\n"))
}

func AddBannedUser(channelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go discord.SayByIDAndDelete(channelID, "No users provided.")
		return
	}

	for _, userToAdd := range args {
		exists := false
		for _, userExisting := range global.BannedUsers {
			if userExisting == userToAdd {
				go discord.SayByIDAndDelete(channelID, userToAdd+" is already added.")
				exists = true
			}
		}

		if !exists {
			global.BannedUsers = append(global.BannedUsers, userToAdd)
		}
	}

	err := global.SaveBannedUsers()
	if err != nil {
		go discord.SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go discord.SayByIDAndDelete(channelID, "Banned users successfully updated.")
	}
}

func RemoveBannedUser(channelID string, messageID string, args []string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)

	if len(args) == 0 {
		go discord.SayByIDAndDelete(channelID, "No regex provided.")
		return
	}

	for _, userToRemove := range args {
		exists := false
		for i, userExisting := range global.BannedUsers {
			if userToRemove == userExisting {
				global.BannedUsers = global.FastRemove(global.BannedUsers, i)
				exists = true
				break
			}
		}

		if !exists {
			go discord.SayByIDAndDelete(channelID, userToRemove+" is not on the list.")
		}
	}

	err := global.SaveBannedUsers()
	if err != nil {
		go discord.SayByIDAndDelete(channelID, "Error:\n"+err.Error())
	} else {
		go discord.SayByIDAndDelete(channelID, "Banned users successfully updated.")
	}
}

func SendBannedUsers(channelID string, messageID string) {
	defer discord.DeleteDiscordMessage(channelID, messageID)
	discord.SayByIDAndDelete(channelID, strings.Join(global.BannedUsers, ",\n"))
}

// UpdateResourceAndChannel updates a resource and the correlated Discord channel.
//
// mode = "add" || "remove"
// func UpdateResourceAndChannel(resourceType string, mode string, returnChannelID string, messageID string, entries []string) {
// 	defer discord.DeleteDiscordMessage(returnChannelID, messageID)

// 	for i := 0; i < len(global.Resources); i++ {
// 		if resourceType == global.Resources[i].DiscordChannelName {
// 			switch mode {

// 			case "add":
// 				for _, entry := range entries {
// 					if global.Resources[i].Content == "" {
// 						global.Resources[i].Content = entry + " "
// 					} else {
// 						global.Resources[i].Content = global.Resources[i].Content + entry + " "
// 					}
// 				}
// 			case "remove":
// 				for _, entry := range entries {
// 					global.Resources[i].Content = strings.ReplaceAll(global.Resources[i].Content, entry+" ", "")
// 				}
// 			}

// 			_, ok := discord.UpdateResourceChannel(global.Resources[i])
// 			if !ok {
// 				go discord.SayByIDAndDelete(returnChannelID, "Failed to update "+resourceType+".")
// 			} else {
// 				go discord.SayByIDAndDelete(returnChannelID, "Successfully updated "+resourceType+".")
// 			}
// 		}
// 	}
// 	global.UpdateResourceLists()
// }

func removeDeterminers(content string) (target string) {
	s := strings.Split(clearNonAlphanumeric(content), " ")
	ns := []string{}
	wordsToAvoid := global.BotName + "^this$|^that$|^those$|^to$|^you$|^i$|^is$|^a$|^the$|^a$|^an$|^this$|^that$|^these$|^those$|^my$|^your$|^his$|^her$|^its$|^our$|^their$|^much$|^many$|^of$|^most$|^some$|^any$|^enough$|^all$|^both$|^half$|^either$|^neither$|^each$|^every$|^other$|^another$|^such$|^rather$|^quite$|^from$"

	for _, w := range s {
		match, err := regexp.MatchString(wordsToAvoid, w)
		if err != nil {
			panic(err)
		}

		if !match {
			ns = append(ns, w)
		}
	}

	if len(ns) == 0 {
		return ""
	}

	return global.PickRandomFromSlice(ns)
}

func clearNonAlphanumeric(str string) string {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func questionType(content string) (questionType string) {
	yesNoWords := []string{"will", "is", "does", "do", "are", "have"}
	explanationWords := []string{"why", "how"}

	for _, q := range yesNoWords {
		if strings.HasPrefix(content, q) {
			return "yes no question"
		}
	}

	for _, q := range explanationWords {
		if strings.HasPrefix(content, q) {
			return "explanation question"
		}
	}

	return "not a question"
}

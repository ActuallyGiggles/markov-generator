package handlers

import (
	"markov-generator/global"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/stats"
	"regexp"
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

func lockResponse(timer int, channel string) bool {
	respondLockMx.Lock()
	if respondLock[channel] {
		respondLockMx.Unlock()
		return false
	}
	respondLock[channel] = true
	respondLockMx.Unlock()
	go unlockResponse(timer, channel)
	return true
}

func unlockResponse(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Second)
	respondLockMx.Lock()
	respondLock[channel] = false
	respondLockMx.Unlock()
}

func GetRandomChannel(channel string) (randomChannel string) {
	// Get a random channel global.Channels list except the channel included and empty channels

	var s []string

	chains := markov.CurrentChains()
	for _, chain := range chains {
		if chain == channel {
			continue
		}
		s = append(s, chain)
	}

	return global.PickRandomFromSlice(s)
}

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

	go discord.SayByID(returnChannelID, "Gathering emotes for "+channelName)
	twitch.GetLiveStatuses()
	ok := twitch.GetEmoteController(false, channel)
	if ok {
		twitch.Join(channelName)
		go discord.SayByID(returnChannelID, channelName+" added successfully.")
	} else {
		go discord.SayByID(returnChannelID, "Could not retrieve "+channelName+"'s emotes...")
		discord.DeleteDiscordChannel(channelName)
	}
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
		stats.Log("failed to remove directive " + channel.ChannelName)
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
		stats.Log("failed to find existing args for " + channelName)
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
		stats.Log("failed to remove directive " + channel.ChannelName)
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

func removeDeterminers(content string) (target string) {
	s := strings.Split(clearNonAlphanumeric(content), " ")
	wordsToAvoid := global.BotName + "|" + "the|a|an|this|that|these|those|my|your|his|her|its|our|their|a few|a little|much|many|a lot of|most|some|any|enough|all|both|half|either|neither|each|every|other|another|such|what|rather|quite"

	for true {
		matched := true

		for i, w := range s {
			match, err := regexp.MatchString(wordsToAvoid, w)
			if err != nil {
				panic(err)
			}

			if match {
				s = global.FastRemove(s, i)
				break
			}

			matched = false
		}

		if !matched {
			break
		}
	}

	return global.PickRandomFromSlice(s)
}

func clearNonAlphanumeric(str string) string {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

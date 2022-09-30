package handler

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"regexp"
	"strings"
	"time"
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
	global.ChannelEmotesMx.Lock()
	defer global.ChannelEmotesMx.Unlock()
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
		for _, emote := range global.ThirdPartyChannelEmotes[channel] {
			if word == emote.Name {
				match = true
				new = append(new, word)
				break
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

func randomlyPickLongerSentences(sentence string) bool {
	// Split sentence into words
	s := strings.Split(sentence, " ")

	// If s does not exceed 1 word, give 25% chance of making it through
	if !DoesSliceContainIndex(s, 1) {
		n := global.RandomNumber(0, 100)
		if n <= 50 {
			return false
		}
	}

	// If s does not exceed 2 words, give 50% chance of making it through
	if !DoesSliceContainIndex(s, 2) {
		n := global.RandomNumber(0, 100)
		if n <= 50 {
			return false
		}
	}

	// If s does not exceed 3 words, give 75% chance of making it through
	if !DoesSliceContainIndex(s, 3) {
		n := global.RandomNumber(0, 100)
		if n <= 75 {
			return false
		}
	}

	return true
}

func DoesSliceContainIndex(slice []string, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

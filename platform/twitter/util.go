package twitter

import "markov-generator/global"

func PickRandomFromMap(potentialTweets map[string]string) (string, string, bool) {
	// Get slice of channels in map
	var channels []string
	for channel := range potentialTweets {
		channels = append(channels, channel)
	}

	if len(channels) == 0 {
		return "", "", true
	}

	// Get random channel
	channel := global.PickRandomFromSlice(channels)

	// Get slice of messages from channel in map
	var messages []string
	for c, message := range potentialTweets {
		if c == channel {
			messages = append(messages, message)
		}
	}

	// Get random message from channel
	message := global.PickRandomFromSlice(messages)

	return channel, message, false
}

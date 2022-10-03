package twitter

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform/twitch"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

var client *gotwi.Client
var potentialTweets = make(map[string]string)
var potentialTweetsMx sync.Mutex

func Start() {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           global.TwitterAccessToken,
		OAuthTokenSecret:     global.TwitterAccessTokenSecret,
	}

	c, err := gotwi.NewClient(in)
	client = c
	if err != nil {
		panic(fmt.Sprintf("Twitter not started.\n %e", err))
	}

	go pickTweet()

	log.Println("Twitter started")
}

func SendTweet(channel string, message string) {
	for _, d := range twitch.Broadcasters {
		if d.Login == channel {
			channel = d.DisplayName
		}
	}

	message = fmt.Sprintf("%s\n\n#%sChatSays\n#ShitTwitchChatSays", channel, message)

	log.Println(fmt.Sprintf("Tweet: \n\tChannel: %s \n\tMessage: %s", channel, strings.ReplaceAll(message, "ChatSays \n", "ChatSays ")))

	p := &types.CreateInput{
		Text: gotwi.String(message),
	}

	_, err := managetweet.Create(context.Background(), client, p)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func AddMessageToPotentialTweets(channel string, message string) {
	// Add to map
	if len(message) > 227 {
		return
	}
	potentialTweetsMx.Lock()
	defer potentialTweetsMx.Unlock()
	potentialTweets[channel] = message
}

func pickTweet() {
	// Create ticker to repeat tweet picking
	for range time.Tick(30 * time.Minute) {
		channel, message, empty := PickRandomFromMap(potentialTweets)
		if empty {
			log.Println("Empty map.")
		} else {
			SendTweet(channel, message)
		}
		potentialTweetsMx.Lock()
		potentialTweets = make(map[string]string)
		potentialTweetsMx.Unlock()
	}
}

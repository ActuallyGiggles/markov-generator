package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitter"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"MarkovGenerator/markov"
)

var (
	IsLive        = make(map[string]bool)
	IsLiveMx      sync.Mutex
	respondLock   = make(map[string]bool)
	respondLockMx sync.Mutex
)

func sendBackOutput(msg platform.Message) {
	for _, directive := range global.Directives {
		if directive.ChannelName == msg.ChannelName {
			onlineEnabled := directive.Settings.IsOnlineEnabled
			offlineEnabled := directive.Settings.IsOfflineEnabled
			random := directive.Settings.IsOptedIn

			IsLiveMx.Lock()
			live := IsLive[directive.ChannelName]
			IsLiveMx.Unlock()

			if (live && onlineEnabled) || (!live && offlineEnabled) {
				if !lockResponse(global.RandomNumber(30, 180), directive.ChannelName) {
					return
				}

				chainToUse := directive.ChannelName

				if random {
					chainToUse, _ = GetRandomChannel(directive.ChannelName)
				}

				if _, ok := jsonToChain("./markov/chains/" + chainToUse + ".json"); !ok {
					log.Println(chainToUse, "does not have a chain yet")
					return
				}

				s := strings.Split(msg.Content, " ")
				t := global.PickRandomFromSlice(s)

				oi := markov.OutputInstructions{
					Method: "TargetedBeginning",
					Chain:  chainToUse,
					Target: t,
				}

				output, problem := markov.Output(oi)
				if problem != "" {
					discord.Say("error-tracking", problem)
				} else {
					discord.Say("all", output)
					discord.Say(directive.ChannelName, output)

					if global.Regex.MatchString(output) {
						discord.Say("quarantine", output)
					} else {
						twitter.AddMessageToPotentialTweets(directive.ChannelName, output)
						Say(directive.ChannelName, output)
					}
				}
				return
			}
		}
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

func GetRandomChannel(channel string) (string, bool) {
	// Get a random channel global.Channels list except the channel included and empty channels

	var s []string

	for _, ch := range global.Directives {
		// Don't choose self or empty channel
		if ch.ChannelName == channel {
			continue
		}
		if _, ok := jsonToChain("./markovdb/" + channel + ".json"); !ok {
			continue
		}
		s = append(s, ch.ChannelName)
	}

	// Check to make sure slice has things in it
	if len(s) == 0 {
		return "", false
	}

	return global.PickRandomFromSlice(s), true
}

func jsonToChain(path string) (chain map[string]map[string]map[string]int, exists bool) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed reading file: %s", err)
		return nil, false
	}

	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("jsonToChain error: ", path, "\n", err)
		return nil, false
	}

	err = json.Unmarshal(content, &chain)
	if err != nil {
		log.Println("Error when unmarshalling file:", path, "\n", err)
		return nil, false
	}

	return chain, true
}

package api

import (
	"encoding/json"
	"log"
	"markov-generator/global"
	"markov-generator/handlers"
	"markov-generator/platform"
	"markov-generator/platform/twitch"
	"markov-generator/stats"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"markov-generator/markov"
)

var (
	in           chan platform.Message
	websiteHits  int
	sentenceHits int
)

func HandleRequests() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/tracked-channels", trackedChannels)
	mux.HandleFunc("/live-channels", liveChannels)
	mux.HandleFunc("/tracked-emotes", trackedEmotes)
	mux.HandleFunc("/get-sentence", getSentence)
	mux.HandleFunc("/server-stats", serverStats)

	//handler := cors.AllowAll().Handler(mux)
	http.ListenAndServe(":10000", mux)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("Hit homePage Endpoint")

	if limitEndpoint(5, "homePage") {
		welcome := struct {
			Welcome string `json:"welcome"`
			Usage   string `json:"usage"`
			Example string `json:"example"`
			PS      string `json:"ps"`
			Socials struct {
				Website string `json:"website"`
				Twitter string `json:"twitter"`
				Discord string `json:"discord"`
				GitHub  string `json:"github"`
			} `json:"socials"`
			TrackedChannelsEndpoint string `json:"tracked_channels_endpoint"`
			TrackedEmotesEndpoint   string `json:"tracked_emotes_endpoint"`
		}{}
		welcome.Welcome = "Welcome to the HomePage!"
		welcome.Usage = "Start using this API by going to /getsentence and ?channel=[channel]"
		welcome.Example = "https://actuallygiggles.localtonet.com/get-sentence?channel=39daph"
		welcome.PS = "Not every channel is being tracked! If you have a suggestion on which channel should be tracked, @ me on Twitter or join the Discord!"
		welcome.Socials.Website = "https://actuallygiggles.github.io/twitch-message-generator/"
		welcome.Socials.Twitter = "https://twitter.com/shit_chat_says"
		welcome.Socials.Discord = "discord.gg/wA96rfyn9p"
		welcome.Socials.GitHub = "https://github.com/ActuallyGiggles/markov-generator"
		welcome.TrackedChannelsEndpoint = "https://actuallygiggles.localtonet.com/tracked-channels"
		welcome.TrackedEmotesEndpoint = "https://actuallygiggles.localtonet.com/tracked-emotes"
		json.NewEncoder(w).Encode(welcome)
	} else {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 5 seconds"
		json.NewEncoder(w).Encode(err)
	}
}

func serverStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("Hit getServerStats Endpoint")

	if limitEndpoint(1, "serverStats") {
		access := r.URL.Query().Get("access")

		if access != "security-omegalul" {
			err := struct {
				Error string
			}{}
			err.Error = "Incorrect security code"
			json.NewEncoder(w).Encode(err)
			return
		}

		s := stats.GetStats()

		s.WebsiteHits = websiteHits
		s.SentenceHits = sentenceHits

		json.NewEncoder(w).Encode(s)
	} else {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 1 second"
		json.NewEncoder(w).Encode(err)
	}
}

func trackedChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("Hit trackedChannels Endpoint")

	if limitEndpoint(5, "trackedChannels") {
		var channels []twitch.Data
		chains := markov.CurrentChains()
		for _, d := range twitch.Broadcasters {
			for _, chain := range chains {
				if d.Login == chain {
					channels = append(channels, d)
				}
			}
		}
		json.NewEncoder(w).Encode(channels)
	} else {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 5 seconds"
		json.NewEncoder(w).Encode(err)
	}
}

func trackedEmotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("Hit trackedEmotes Endpoint")

	if limitEndpoint(5, "trackedEmotes") {
		allEmotes := struct {
			Global []global.Emote `json:"global"`
		}{}
		allEmotes.Global = append(allEmotes.Global, global.GlobalEmotes...)
		allEmotes.Global = append(allEmotes.Global, global.TwitchChannelEmotes...)
		json.NewEncoder(w).Encode(allEmotes)
	} else {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 5 seconds"
		json.NewEncoder(w).Encode(err)
	}
}

func liveChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("Hit liveChannels Endpoint")
	websiteHits += 1

	if limitEndpoint(5, "liveChannels") {
		var channelsLive []struct {
			Name string
			Live bool
		}
		for channel, status := range twitch.IsLive {
			e := struct {
				Name string
				Live bool
			}{}
			e.Name = channel
			e.Live = status
			channelsLive = append(channelsLive, e)
		}
		json.NewEncoder(w).Encode(channelsLive)
	} else {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 5 seconds"
		json.NewEncoder(w).Encode(err)
	}
}

func getSentence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("Hit getSentence Endpoint")
	sentenceHits += 1

	channel := strings.ToLower(r.URL.Query().Get("channel"))
	//method := r.URL.Query().Get("method")
	//target := r.URL.Query().Get("target")

	var (
		apiResponse   APIResponse
		instructions  = markov.OutputInstructions{Chain: channel, Method: global.PickRandomFromSlice([]string{"LikelyBeginning", "LikelyEnding"})}
		timesRecursed int
		output        string
		err           error
	)

	if !lockChannel(.5, channel) {
		apiResponse.Error = "Channel is locked for 0.5s to prevent spam. Try again."
		goto reply
	}

recurse:
	output, err = markov.Out(instructions)
	apiResponse.MarkovSentence = output

	if err != nil || output == "" || handlers.IsSentenceFiltered(output) {
		if timesRecursed < 100 {
			timesRecursed++
			goto recurse
		}
		apiResponse.Error = "There is a server side issue."
		goto reply
	}

reply:
	handlers.OutputHandler("api", channel, channel, output, "")
	json.NewEncoder(w).Encode(apiResponse)
	return
}

// func getTwitchBroadcasterInfo(w http.ResponseWriter, r *http.Request) {
// 	log.Println("Hit getTwitchBroadcasterInfo Endpoint")
// 	w.Header().Set("Content-Type", "application/json")
// 	channel := strings.ToLower(r.URL.Query().Get("channel"))

// 	response := struct {
// 		ID              string `json:"id"`
// 		Login           string `json:"login"`
// 		DisplayName     string `json:"display_name"`
// 		Type            string `json:"type"`
// 		BroadcasterType string `json:"broadcaster_type"`
// 		Description     string `json:"description"`
// 		ProfileImageUrl string `json:"profile_image_url"`
// 		OfflineImageUrl string `json:"offline_image_url"`
// 		ViewCount       int    `json:"view_count"`
// 		Email           string `json:"email"`
// 		CreatedAt       string `json:"created_at"`
// 		Error           string `json:"error"`
// 	}{}

// 	if channel == "" {
// 		response.Error = "Channel is blank! Append ?channel=[channelname] to the url."
// 	}

// 	d, err := twitch.GetBroadcasterInfo(channel)
// 	if err != nil {
// 		response.Error = "Something went wrong... Is this a real user? Are they banned?"
// 	} else {
// 		response.ID = d.ID
// 		response.Login = d.Login
// 		response.DisplayName = d.DisplayName
// 		response.Type = d.Type
// 		response.BroadcasterType = d.BroadcasterType
// 		response.Description = d.Description
// 		response.ProfileImageUrl = d.ProfileImageUrl
// 		response.OfflineImageUrl = d.OfflineImageUrl
// 		response.ViewCount = d.ViewCount
// 		response.Email = d.Email
// 		response.CreatedAt = d.CreatedAt
// 		response.Error = ""
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

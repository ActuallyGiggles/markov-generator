package api

import (
	"encoding/json"
	"markov-generator/global"
	"markov-generator/handlers"
	"markov-generator/platform"
	"markov-generator/platform/twitch"
	"markov-generator/print"
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
	mux.HandleFunc("/", apiHomePage)

	mux.HandleFunc("/data", websiteHomePage)
	mux.HandleFunc("/emotes", emotes)

	mux.HandleFunc("/get-sentence", getSentence)

	mux.HandleFunc("/server-stats", serverStats)

	//handler := cors.AllowAll().Handler(mux)
	http.ListenAndServe(":10000", mux)
}

func apiHomePage(w http.ResponseWriter, r *http.Request) {
	print.Info("Hit API Home Page")

	w.Header().Set("Content-Type", "application/json")

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

func websiteHomePage(w http.ResponseWriter, r *http.Request) {
	print.Info("Hit Website Home Page")
	websiteHits++

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if !limitEndpoint(5, "website home page") {
		err := struct {
			Error string
		}{}
		err.Error = "Endpoint Limiter: Try again in 5 seconds"
		json.NewEncoder(w).Encode(err)
	}

	var data DataSend

	// Channels Used
	chains := markov.CurrentWorkers()
	for _, d := range twitch.Broadcasters {
		for _, chain := range chains {
			if d.Login == chain {
				data.ChannelsUsed = append(data.ChannelsUsed, d)
			}
		}
	}

	// Channels Live
	for channel, status := range twitch.IsLive {
		e := struct {
			Name string
			Live bool
		}{}
		e.Name = channel
		e.Live = status
		data.ChannelsLive = append(data.ChannelsLive, e)
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		panic(err)
	}
}

func emotes(w http.ResponseWriter, r *http.Request) {
	print.Info("Hit Emotes Endpoint")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

func getSentence(w http.ResponseWriter, r *http.Request) {
	print.Info("Hit Sentence Endpoint")
	sentenceHits++

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

func serverStats(w http.ResponseWriter, r *http.Request) {
	print.Info("Hit Stats Endpoint")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

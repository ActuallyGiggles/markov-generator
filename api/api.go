package api

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitch"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"MarkovGenerator/markov"

	"github.com/rs/cors"
)

var in chan platform.Message

func homePage(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit homePage Endpoint")
	w.Header().Set("Content-Type", "application/json")
	welcome := struct {
		Welcome string `json:"welcome"`
		Usage   string `json:"usage"`
		Example string `json:"example"`
		PS      string `json:"ps"`
		Socials struct {
			Twitter string `json:"twitter"`
			Discord string `json:"discord"`
			GitHub  string `json:"github"`
		} `json:"socials"`
		ChannelsTracked []string `json:"channels-tracked"`
	}{}

	welcome.Welcome = "Welcome to the HomePage!"
	welcome.Usage = "Start using this API by going to /getsentence and ?channel=[channel]"
	welcome.Example = "https://actuallygiggles.localtonet.com/getsentence?channel=39daph"
	welcome.PS = "Not every channel is being tracked! If you have a suggestion on which channel should be tracked, @ me on Twitter or join the Discord!"
	welcome.Socials.Twitter = "https://twitter.com/shit_chat_says"
	welcome.Socials.Discord = "discord.gg/wA96rfyn9p"
	welcome.Socials.GitHub = "https://github.com/ActuallyGiggles/markov-generator"
	welcome.ChannelsTracked = markov.Chains()

	json.NewEncoder(w).Encode(welcome)
}

func getSentence(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit getSentence Endpoint")
	w.Header().Set("Content-Type", "application/json")
	method := r.URL.Query().Get("method")
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	target := r.URL.Query().Get("target")

	var apiResponse APIResponse

	if method == "" {
		method = "beginning"
	}

	if channel == "" {
		apiResponse = APIResponse{
			ModeUsed:       method,
			ChannelUsed:    channel,
			MessageUsed:    target,
			MarkovSentence: "",
			Error:          "Channel is blank!",
		}
	} else {
		exists := false
		for _, directive := range global.Directives {
			if directive.ChannelName == channel {
				exists = true
				break
			}
		}
		if !exists {
			apiResponse = APIResponse{
				ModeUsed:       method,
				ChannelUsed:    channel,
				MessageUsed:    target,
				MarkovSentence: "",
				Error:          strings.Title(channel) + " is not being tracked!",
			}
		} else {
			if lockChannel(1, channel) {
				oi := markov.OutputInstructions{
					Method: "LikelyBeginning",
					Chain:  channel,
				}
				output, problem := markov.Output(oi)
				if problem == "" {
					apiResponse = APIResponse{
						ModeUsed:       method,
						ChannelUsed:    channel,
						MessageUsed:    target,
						MarkovSentence: output,
						Error:          problem,
					}

					m := platform.Message{
						Platform:    "api",
						ChannelName: channel,
						Content:     output,
					}

					in <- m
				} else {
					discord.Say("error-tracking", problem)
				}
			} else {
				apiResponse = APIResponse{
					ModeUsed:       method,
					ChannelUsed:    channel,
					MessageUsed:    target,
					MarkovSentence: "",
					Error:          "channel is locked for 1s to prevent spam",
				}
			}
		}
	}

	json.NewEncoder(w).Encode(apiResponse)
}

func getTwitchBroadcasterInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit getTwitchBroadcasterInfo Endpoint")
	w.Header().Set("Content-Type", "application/json")
	channel := strings.ToLower(r.URL.Query().Get("channel"))

	response := struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Type            string `json:"type"`
		BroadcasterType string `json:"broadcaster_type"`
		Description     string `json:"description"`
		ProfileImageUrl string `json:"profile_image_url"`
		OfflineImageUrl string `json:"offline_image_url"`
		ViewCount       int    `json:"view_count"`
		Email           string `json:"email"`
		CreatedAt       string `json:"created_at"`
		Error           string `json:"error"`
	}{}

	d, ok := twitch.GetBroadcasterInfo(channel)
	if !ok {
		response.Error = "Something went wrong... Is this a real user? Are they banned?"
	} else {
		response.ID = d.ID
		response.Login = d.Login
		response.DisplayName = d.DisplayName
		response.Type = d.Type
		response.BroadcasterType = d.BroadcasterType
		response.Description = d.Description
		response.ProfileImageUrl = d.ProfileImageUrl
		response.OfflineImageUrl = d.OfflineImageUrl
		response.ViewCount = d.ViewCount
		response.Email = d.Email
		response.CreatedAt = d.CreatedAt
		response.Error = ""
	}

	json.NewEncoder(w).Encode(response)
}

func HandleRequests(c chan platform.Message) {
	in = c
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/get-sentence", getSentence)
	mux.HandleFunc("/twitch-broadcaster-info", getSentence)

	handler := cors.Default().Handler(mux)
	log.Println("API started")
	log.Fatal(http.ListenAndServe(":10000", handler))
}

type APIResponse struct {
	ModeUsed       string `json:"mode_used"`
	ChannelUsed    string `json:"channel_used"`
	MessageUsed    string `json:"message_used"`
	MarkovSentence string `json:"markov_sentence"`
	Error          string `json:"error"`
}

package api

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform/discord"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ActuallyGiggles/go-markov"
	"github.com/rs/cors"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	welcome := struct {
		Welcome string
		Usage   string
		Example string
		PS      string
		Socials struct {
			Twitter string
			Discord string
			GitHub  string
		}
	}{
		Welcome: "Welcome to the HomePage!",
		Usage:   "Start using this API by going to /getsentence and ?channel=[channel]",
		Example: "https://actuallygiggles.localtonet.com/getsentence?channel=39daph",
		PS:      "Not every channel is being tracked! If you have a suggestion on which channel should be tracked, @ me on Twitter or join the Discord!",
		Socials: struct {
			Twitter string
			Discord string
			GitHub  string
		}{
			Twitter: "https://twitter.com/shit_chat_says",
			Discord: "discord.gg/wA96rfyn9p",
			GitHub:  "https://github.com/ActuallyGiggles/markov-generator",
		},
	}
	fmt.Fprintf(w, "Welcome to the HomePage!")
	json.NewEncoder(w).Encode(welcome)
}

func getSentencePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mode := r.URL.Query().Get("mode")
	channel := r.URL.Query().Get("channel")
	message := r.URL.Query().Get("message")

	var apiResponse APIResponse

	if mode == "" {
		mode = "beginning"
	}

	if channel == "" {
		apiResponse = APIResponse{
			ModeUsed:       mode,
			ChannelUsed:    channel,
			MessageUsed:    message,
			MarkovSentence: "",
			Error:          "Channel is blank!",
		}
	} else {
		// _, ok := twitch.GetBroadcasterID(channel)
		// if !ok {
		// 	apiResponse = APIResponse{
		// 		ModeUsed:       mode,
		// 		ChannelUsed:    channel,
		// 		MessageUsed:    message,
		// 		MarkovSentence: "",
		// 		Error:          strings.Title(channel) + " is not real!",
		// 	}
		// } else {
		exists := false
		for _, directive := range global.Directives {
			if directive.ChannelName == channel {
				exists = true
				break
			}
		}
		if !exists {
			apiResponse = APIResponse{
				ModeUsed:       mode,
				ChannelUsed:    channel,
				MessageUsed:    message,
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
						ModeUsed:       mode,
						ChannelUsed:    channel,
						MessageUsed:    message,
						MarkovSentence: output,
						Error:          problem,
					}
				} else {
					discord.Say("error-tracking", problem)
				}
			} else {
				apiResponse = APIResponse{
					ModeUsed:       mode,
					ChannelUsed:    channel,
					MessageUsed:    message,
					MarkovSentence: "",
					Error:          "channel is locked for 1s to prevent spam",
				}
			}
		}
		// }
	}

	json.NewEncoder(w).Encode(apiResponse)
}

func HandleRequests() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/getsentence", getSentencePage)

	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":10000", handler))
}

type APIResponse struct {
	ModeUsed       string `json:"mode_used"`
	ChannelUsed    string `json:"channel_used"`
	MessageUsed    string `json:"message_used"`
	MarkovSentence string `json:"markov_sentence"`
	Error          string `json:"error"`
}

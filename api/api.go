package api

import (
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ActuallyGiggles/go-markov"
	"github.com/rs/cors"
)

var in chan platform.Message

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
		ChannelsTracked []string
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
		ChannelsTracked: markov.Chains(),
	}
	json.NewEncoder(w).Encode(welcome)
}

func getSentencePage(w http.ResponseWriter, r *http.Request) {
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

func HandleRequests(c chan platform.Message) {
	in = c
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/getsentence", getSentencePage)

	handler := cors.Default().Handler(mux)
	fmt.Println("API started")
	log.Fatal(http.ListenAndServe(":10000", handler))
}

type APIResponse struct {
	ModeUsed       string `json:"mode_used"`
	ChannelUsed    string `json:"channel_used"`
	MessageUsed    string `json:"message_used"`
	MarkovSentence string `json:"markov_sentence"`
	Error          string `json:"error"`
}

package twitch

import (
	"MarkovGenerator/platform"
	"fmt"
	"time"
)

var (
	didInitializationHappen = false
)

func Start(in chan platform.Message) {
	msg := <-in

	if msg.Platform == "internal" && msg.Content == "finished discord init" {
		fmt.Println("Discord Started")
	}

	fmt.Println("Fetching live statuses...")
	GetLiveStatuses()

	fmt.Println("Fetching emotes...")
	getEmoteController()

	didInitializationHappen = true

	clientStart(in)

	fmt.Println("Twitch Started")
	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(30 * time.Second) {
		fmt.Println("Updating live statuses...")
		GetLiveStatuses()
	}
}

func refreshEmotes() {
	for range time.Tick(10 * time.Minute) {
		getEmoteController()
	}
}

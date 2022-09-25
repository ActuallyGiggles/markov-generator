package twitch

import (
	"fmt"
	"time"
)

var (
	didInitializationHappen = false
)

func GatherEmotes() {
	fmt.Println("Gathering emotes")
	GetLiveStatuses()
	getEmoteController()
	didInitializationHappen = true
	fmt.Println("Emotes gathered")
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

package twitch

import (
	"log"
	"time"
)

var (
	didInitializationHappen = false
)

func GatherEmotes() {
	log.Println("Gathering emotes")
	GetLiveStatuses()
	GetEmoteController()
	didInitializationHappen = true
	log.Println("Emotes gathered")
	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(30 * time.Second) {
		log.Println("Updating live statuses...")
		GetLiveStatuses()
	}
}

func refreshEmotes() {
	for range time.Tick(10 * time.Minute) {
		GetEmoteController()
	}
}

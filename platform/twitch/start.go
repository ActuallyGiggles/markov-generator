package twitch

import (
	"log"
	"sync"
	"time"
)

var (
	IsLive   = make(map[string]bool)
	IsLiveMx sync.Mutex
)

func GatherEmotes() {
	log.Println("Gathering emotes")
	GetLiveStatuses()
	GetEmoteController(true)
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
		GetEmoteController(false)
	}
}

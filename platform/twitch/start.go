package twitch

import (
	"markov-generator/stats"
	"sync"
	"time"
)

var (
	IsLive   = make(map[string]bool)
	IsLiveMx sync.Mutex
)

func GatherEmotes() {
	stats.Log("Gathering emotes")
	GetLiveStatuses()
	GetEmoteController(true)
	stats.Log("Emotes gathered")
	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(30 * time.Second) {
		stats.Log("Updating live statuses...")
		GetLiveStatuses()
	}
}

func refreshEmotes() {
	for range time.Tick(10 * time.Minute) {
		GetEmoteController(false)
	}
}

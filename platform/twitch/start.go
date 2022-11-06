package twitch

import (
	"sync"
	"time"
)

var (
	IsLive   = make(map[string]bool)
	IsLiveMx sync.Mutex
)

func GatherEmotes() {
	GetLiveStatuses()
	GetEmoteController(true)
	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(2 * time.Minute) {
		GetLiveStatuses()
	}
}

func refreshEmotes() {
	for range time.Tick(10 * time.Minute) {
		GetEmoteController(false)
	}
}

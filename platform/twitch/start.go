package twitch

import (
	"markov-generator/global"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

var (
	IsLive   = make(map[string]bool)
	IsLiveMx sync.Mutex
	bar      *progressbar.ProgressBar
)

func GatherEmotes() {
	bar = progressbar.Default(int64(4+len(global.Directives)*6), "Collecting Twitch API information...")
	GetLiveStatuses(true)
	GetEmoteController(true, global.Directive{})
	bar.Clear()

	go updateLiveStatuses()
	go refreshEmotes()
}

func updateLiveStatuses() {
	for range time.Tick(2 * time.Minute) {
		GetLiveStatuses(false)
	}
}

func refreshEmotes() {
	for range time.Tick(30 * time.Minute) {
		GetEmoteController(false, global.Directive{})
	}
}

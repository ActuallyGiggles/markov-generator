package main

import (
	"io"
	"log"
	"markov-generator/api"
	"markov-generator/global"
	"markov-generator/handler"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"markov-generator/stats"

	"sync"

	"os"
	"os/signal"
	"syscall"

	"markov-generator/markov"

	"github.com/pkg/profile"
)

func main() {
	// Profiling
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	// Logging
	file, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	wrt := io.MultiWriter(os.Stdout, file)
	log.SetOutput(wrt)

	// Keep open
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	Start()

	<-sc
	stats.Log("Stopping...")
}

func Start() {
	stats.Log("Initializing")

	c := make(chan platform.Message)

	global.Start()
	stats.Log("Global started")

	go twitter.Start()
	stats.Log("Twitter started")

	go handler.MsgHandler(c)
	stats.Log("Handler started")

	var wg sync.WaitGroup
	wg.Add(1)
	go discord.Start(c, &wg)
	stats.Log("Discord started")
	wg.Wait()

	stats.Log("Gathering emotes")
	twitch.GatherEmotes()
	stats.Log("Gathered emotes")

	i := markov.StartInstructions{
		WriteMode:  "counter",
		WriteLimit: 10000,
		StartKey:   "b5G(n1$I!4g",
		EndKey:     "e1$D(n7",
		Debug:      true,
	}
	markov.Start(i)
	stats.Log("Markov started")

	go api.HandleRequests()
	stats.Log("API started")

	go twitch.Start(c)
	stats.Log("Twitch Started")

	stats.Start()

	stats.Log("Initialization complete")
}

package main

import (
	"fmt"
	"io"
	"log"
	"markov-generator/api"
	"markov-generator/global"
	"markov-generator/handlers"
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
	file, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE, 0666)
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
	c := make(chan platform.Message)

	global.Start()

	go twitter.Start()
	fmt.Println("✅ Twitter started")

	var wg sync.WaitGroup
	wg.Add(1)
	go discord.Start(c, &wg)
	wg.Wait()
	fmt.Println("✅ Discord started")

	i := markov.StartInstructions{
		WriteMode:     "interval",
		WriteInterval: 10,
		IntervalUnit:  "minutes",
		// WriteMode:  "counter",
		// WriteLimit: 10,
		StartKey: "b5G(n1$I!4g",
		EndKey:   "e1$D(n7",
		Debug:    false,
	}
	markov.Start(i)
	fmt.Println("✅ Markov Started")

	go handlers.MsgHandler(c)

	twitch.GatherEmotes()

	go api.HandleRequests()

	go twitch.Start(c)
	fmt.Println("✅ Twitch Started")

	stats.Start()
}

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
	"markov-generator/terminal"

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
	log.Println("Stopping...")
}

func Start() {

	defer global.RecoverFullName("main.go -> Start()")

	c := make(chan platform.Message)

	global.Start()
	log.Println("Global started")

	go twitter.Start()
	go handler.MsgHandler(c)

	var wg sync.WaitGroup
	wg.Add(1)
	go discord.Start(c, &wg)
	wg.Wait()

	twitch.GatherEmotes()

	i := markov.StartInstructions{
		WriteMode: "counter",
		// WriteInterval: 10,
		// IntervalUnit:  "minutes",
		WriteLimit: 1000,
		StartKey:   "b5G(n1$I!4g",
		EndKey:     "e1$D(n7",
		Debug:      false,
	}
	markov.Start(i)
	log.Println("Markov started")

	go api.HandleRequests()

	go twitch.Start(c)

	terminal.UpdateTerminal("init")
}

package main

import (
	"MarkovGenerator/api"
	"MarkovGenerator/global"
	"MarkovGenerator/handler"
	"MarkovGenerator/markov"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitch"
	"MarkovGenerator/platform/twitter"
	"MarkovGenerator/terminal"
	"io"
	"log"
	"sync"

	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/profile"
)

func main() {
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	// Logging
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

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
		WriteInterval: 10,
		IntervalUnit:  "minutes",
		StartKey:      "b5G(n1$I!4g",
		EndKey:        "e1$D(n7",
	}
	markov.Start(i)
	log.Println("Markov started")

	go api.HandleRequests(c)

	go twitch.Start(c)

	terminal.UpdateTerminal("init")
}

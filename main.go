package main

import (
	"MarkovGenerator/api"
	"MarkovGenerator/global"
	"MarkovGenerator/handler"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitch"
	"MarkovGenerator/platform/twitter"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ActuallyGiggles/go-markov"
)

func main() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	global.Start()
	fmt.Println("Global started")

	i := markov.StartInstructions{
		Workers:       10,
		WriteInterval: 10,
		IntervalUnit:  "seconds",
		StartKey:      "b5G(n1$I!4g",
		EndKey:        "e1$D(n7",
	}
	markov.Start(i)
	fmt.Println("Markov started")

	Start()

	<-sc
	fmt.Println("Stopping...")
}

func Start() {
	c := make(chan platform.Message)

	go twitter.Start()
	go api.HandleRequests(c)
	go handler.MsgHandler(c)

	go discord.Start(c)
	msg := <-c
	if msg.Platform == "internal" {
		fmt.Println("Directives gathered")
	}
	twitch.GatherEmotes()
	go twitch.Start(c)
}

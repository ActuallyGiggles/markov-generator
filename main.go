package main

import (
	"context"
	"markov-generator/api"
	"markov-generator/global"
	"markov-generator/handlers"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"markov-generator/print"
	"markov-generator/stats"
	"time"

	"sync"

	"os/signal"
	"syscall"

	"markov-generator/markov"

	"github.com/pkg/profile"
)

func main() {
	// Profiling
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	// Keep open
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	Start()

	go print.TerminalInput(cancel)

	<-ctx.Done()
	print.Page("Exited")
}

func Start() {
	print.Page("Started")

	c := make(chan platform.Message)

	global.Start()

	go twitter.Start()
	print.Success("Twitter")

	var wg sync.WaitGroup
	wg.Add(1)
	go discord.Start(c, &wg)
	wg.Wait()
	print.Success("Discord")

	i := markov.StartInstructions{
		WriteMode:     "interval",
		WriteInterval: 10,
		IntervalUnit:  "minutes",
		// WriteMode:  "counter",
		// WriteLimit: 10,
		StartKey:  "b5G(n1$I!4g",
		EndKey:    "e1$D(n7",
		Debug:     false,
		ShouldZip: true,
	}
	markov.Start(i)
	print.Success("Markov")

	go handlers.MsgHandler(c)

	twitch.GatherEmotes()
	print.Success("Emotes")

	go api.HandleRequests()

	go twitch.Start(c)
	print.Success("Twitch")

	stats.Start()

	print.Page("Twitch Message Generator")
	print.Started("Program Started at " + time.Now().Format(time.RFC822))
}

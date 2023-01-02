package main

import (
	"context"
	"fmt"
	"markov-generator/api"
	"markov-generator/global"
	"markov-generator/handlers"
	"markov-generator/platform"
	"markov-generator/platform/discord"
	"markov-generator/platform/twitch"
	"markov-generator/platform/twitter"
	"markov-generator/print"
	"markov-generator/stats"
	"strings"
	"time"

	"sync"

	"os/signal"
	"syscall"

	"markov-generator/markov"

	"github.com/pkg/profile"
	"github.com/pterm/pterm"
)

func main() {
	// Profiling
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	// Keep open
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	Start()

	<-ctx.Done()
	stats.Log("Stopping...")
}

func Start() {
	print.Page("Twitch Message Generator")

	c := make(chan platform.Message)

	global.Start()

	go twitter.Start()
	print.Success("Twitter")

	var wg sync.WaitGroup
	wg.Add(1)
	go discord.Start(c, &wg)
	wg.Wait()
	print.Success("Discord")

	durationReports := make(chan string)
	go printReports(durationReports)

	i := markov.StartInstructions{
		WriteMode:     "interval",
		WriteInterval: 10,
		IntervalUnit:  "minutes",
		// WriteMode:  "counter",
		// WriteLimit: 10,
		StartKey:        "b5G(n1$I!4g",
		EndKey:          "e1$D(n7",
		ReportDurations: durationReports,
		Debug:           false,
		ShouldZip:       true,
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

	print.Success("Program Started at " + time.Now().Format(time.RFC822))
}

func printReports(reports chan string) {
	for report := range reports {
		split := strings.Split(report, ":")
		process, duration := split[0], split[1]

		pterm.Println()
		print.Success(process + fmt.Sprintf(" \n%s", pterm.Gray(duration)))
		pterm.Println()
	}
}

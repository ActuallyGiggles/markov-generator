package main

import (
	"MarkovGenerator/api"
	"MarkovGenerator/commands"
	"MarkovGenerator/global"
	"MarkovGenerator/platform"
	"MarkovGenerator/platform/discord"
	"MarkovGenerator/platform/twitch"
	"MarkovGenerator/platform/twitter"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ActuallyGiggles/go-markov"
)

func main() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	global.Start()
	fmt.Println("Global Started")

	i := markov.StartInstructions{
		Workers:       10,
		WriteInterval: 10,
		IntervalUnit:  "seconds",
		StartKey:      "b5G(n1$I!4g",
		EndKey:        "e1$D(n7",
	}
	markov.Start(i)
	go outputTicker()
	fmt.Println("Markov Started")

	go twitter.Start()
	fmt.Println("Twitter Started")

	go api.HandleRequests()
	fmt.Println("API Started")

	c := make(chan platform.Message)
	go msgHandler(c)
	go discord.Start(c)
	go twitch.Start(c)

	<-sc
	fmt.Println("Stopping...")
}

func msgHandler(c chan platform.Message) {
	for msg := range c {
		if msg.Platform == "twitch" {
			newMessage, passed := prepareMessage(msg)
			if passed {
				markov.Input(msg.ChannelName, newMessage)
			}
		} else if msg.Platform == "discord" {
			commands.AdminCommands(msg)
		}
	}
}

func outputTicker() {
	for range time.Tick(30 * time.Second) {
		chains := markov.Chains()
		for _, chain := range chains {
			oi := markov.OutputInstructions{
				Method: "LikelyBeginning",
				Chain:  chain,
			}
			output, problem := markov.Output(oi)
			if problem != "" {
				discord.Say("error-tracking", problem)
			} else {
				str := "Channel: " + chain + "\nMessage: " + output
				discord.Say("all", str)
				discord.Say(chain, output)

				if global.Regex.MatchString(output) {
					discord.Say("quarantine", output)
				} else {
					twitter.AddMessageToPotentialTweets(chain, output)
				}
			}
		}
	}
}

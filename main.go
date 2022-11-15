package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	C chan Message

	Users         []User
	GlobalEmotes  []string
	ChannelEmotes []string
)

func main() {
	// Keep open
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	readConfig()

	getEmotes()
	go GetLiveStatuses()

	C := make(chan Message)
	go Start(C)
	go Mimic(C)

	<-sc
	fmt.Println("Stopping...")
}

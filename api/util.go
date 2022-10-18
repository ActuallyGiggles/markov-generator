package api

import (
	"fmt"
	"markov-generator/handler"
	"sync"
	"time"

	"markov-generator/markov"
)

var (
	channelLock   = make(map[string]bool)
	channelLockMx sync.Mutex
	recursions    = make(map[string]int)
	recursionsMx  sync.Mutex
	limit         = make(map[string]int)
	limitMx       sync.Mutex
)

func lockChannel(timer float64, channel string) bool {
	channelLockMx.Lock()
	if channelLock[channel] {
		channelLockMx.Unlock()
		return false
	}
	channelLock[channel] = true
	channelLockMx.Unlock()
	go unlockChannel(timer, channel)
	return true
}

func unlockChannel(timer float64, channel string) {
	time.Sleep(time.Duration(timer * 1e9))
	channelLockMx.Lock()
	channelLock[channel] = false
	channelLockMx.Unlock()
}

func limitEndpoint(timer int, endpoint string) bool {
	limitMx.Lock()
	endpointLimit := 0
	switch endpoint {
	default:
		endpointLimit = 10
	}
	if limit[endpoint] > endpointLimit {
		fmt.Println(endpoint, "not passed", limit[endpoint])
		limitMx.Unlock()
		return false
	}
	limit[endpoint] += 1
	fmt.Println(endpoint, "passed", limit[endpoint])
	limitMx.Unlock()
	go func(timer int) {
		time.Sleep(time.Duration(timer * 1e9))
		limitMx.Lock()
		limit[endpoint] -= 1
		limitMx.Unlock()
		fmt.Println(endpoint, "cleared by 1", limit[endpoint])
	}(timer)
	return true
}

func warden(channel string) (output string) {
	c := make(chan string)
	go guard(channel, c)
	r := <-c
	return r
}

func guard(channel string, c chan string) {
	oi := markov.OutputInstructions{
		Chain:  channel,
		Method: "LikelyBeginning",
	}
	output, problem := markov.Out(oi)

	if problem == nil {
		if !handler.RandomlyPickLongerSentences(output) {
			recurse(channel, output, c)
			return
		} else {
			c <- output
			close(c)
			return
		}
	} else {
		recurse(channel, "", c)
		return
	}
}

func recurse(channel string, output string, c chan string) {
	recursionsMx.Lock()
	recursions[channel] += 1
	if recursions[channel] > 100 {
		recursions[channel] = 0
		recursionsMx.Unlock()
		c <- output
		close(c)
	} else {
		recursionsMx.Unlock()
		go guard(channel, c)
	}
}

package api

import (
	"sync"
	"time"
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
		endpointLimit = 30
	}
	if limit[endpoint] > endpointLimit {
		limitMx.Unlock()
		return false
	}
	limit[endpoint] += 1
	limitMx.Unlock()
	go func(timer int) {
		time.Sleep(time.Duration(timer * 1e9))
		limitMx.Lock()
		limit[endpoint] -= 1
		limitMx.Unlock()
	}(timer)
	return true
}

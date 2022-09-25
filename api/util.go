package api

import (
	"sync"
	"time"
)

var (
	channelLock   = make(map[string]bool)
	channelLockMx sync.Mutex
)

func lockChannel(timer int, channel string) bool {
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

func unlockChannel(timer int, channel string) {
	time.Sleep(time.Duration(timer) * time.Second)
	channelLockMx.Lock()
	channelLock[channel] = false
	channelLockMx.Unlock()
}

package stats

import (
	"markov-generator/markov"
	"time"
)

type Stats struct {
	StartTime      time.Time
	RunTime        time.Duration
	WriteMode      string
	TimeUntilWrite time.Duration
	CurrentCount   int
	CountLimit     int
	PeakIntake     struct {
		Chain  string
		Amount int
		Time   time.Time
	}
}

var StartTime time.Time

func Start() {
	StartTime = time.Now()
}

func GetStats() (stats Stats) {
	stats.StartTime = StartTime
	stats.RunTime = time.Now().Sub(StartTime)
	stats.WriteMode = markov.WriteMode()
	stats.TimeUntilWrite = markov.TimeUntilWrite()
	stats.CurrentCount = markov.CurrentCount
	stats.CountLimit = markov.WriteCountLimit
	stats.PeakIntake = markov.PeakIntake()

	return stats
}

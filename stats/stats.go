package stats

import (
	"markov-generator/markov"
	"runtime"
	"time"
)

type Stats struct {
	StartTime      time.Time
	RunTime        time.Duration
	MemoryUsage    MemoryUsage
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
	stats.MemoryUsage = PrintMemUsage()
	stats.WriteMode = markov.WriteMode()
	stats.TimeUntilWrite = markov.TimeUntilWrite()
	stats.CurrentCount = markov.CurrentCount
	stats.CountLimit = markov.WriteCountLimit
	stats.PeakIntake = markov.PeakIntake()

	return stats
}

type MemoryUsage struct {
	Allocated      uint64
	TotalAllocated uint64
	System         uint64
	NumGC          uint32
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() (mu MemoryUsage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	mu.Allocated = bToMb(m.Alloc)
	mu.TotalAllocated = bToMb(m.TotalAlloc)
	mu.System = bToMb(m.Sys)
	mu.NumGC = m.NumGC

	return mu
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

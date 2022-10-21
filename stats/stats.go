package stats

import (
	"fmt"
	"log"
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
	TotalCount     int
	CurrentCount   int
	CountLimit     int
	Workers        int
	PeakIntake     struct {
		Chain  string
		Amount int
		Time   time.Time
	}
	IntakePerMinute int
	Logs            []string
}

var StartTime time.Time
var IntakePerMinute int
var previousTotal int
var Logs []string

func Start() {
	StartTime = time.Now()

	go intakePerMinute()
}

func Log(message ...string) {
	log.Println(message)
	ct := time.Now()
	year, month, day := ct.Date()
	hour := ct.Hour()
	minute := ct.Minute()
	second := ct.Second()
	Logs = append(Logs, fmt.Sprintf("%d/%d/%d %d:%d:%d %s", year, int(month), day, hour, minute, second, message))
}

func GetStats() (stats Stats) {
	stats.StartTime = StartTime
	stats.RunTime = time.Now().Sub(StartTime)
	stats.MemoryUsage = PrintMemUsage()
	stats.WriteMode = markov.WriteMode()
	stats.TimeUntilWrite = markov.TimeUntilWrite()
	stats.TotalCount = markov.TotalCount
	stats.CurrentCount = markov.CurrentCount
	stats.CountLimit = markov.WriteCountLimit
	stats.Workers = len(markov.CurrentChains())
	stats.PeakIntake = markov.PeakIntake()
	stats.IntakePerMinute = IntakePerMinute
	stats.Logs = Logs

	return stats
}

func intakePerMinute() {
	for range time.Tick(1 * time.Minute) {
		lastMinute := markov.TotalCount - previousTotal
		previousTotal = markov.TotalCount
		IntakePerMinute = lastMinute
	}
}

type MemoryUsage struct {
	Allocated      uint64 `json:"allocated"`
	TotalAllocated uint64 `json:"total_allocated"`
	System         uint64 `json:"system"`
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

	return mu
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

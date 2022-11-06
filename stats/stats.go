package stats

import (
	"fmt"
	"log"
	"markov-generator/markov"
	"runtime"
	"time"
)

type Stats struct {
	StartTime time.Time     `json:"start_time"`
	RunTime   time.Duration `json:"run_time"`

	WriteMode      string                  `json:"write_mode"`
	TimeUntilWrite time.Duration           `json:"time_until_write"`
	CurrentCount   int                     `json:"current_count"`
	CountLimit     int                     `json:"count_limit"`
	IntakePerHour  int                     `json:"intake_per_hour"`
	Workers        int                     `json:"workers"`
	PeakIntake     markov.PeakIntakeStruct `json:"peak_intake"`

	TotalInputs  int `json:"total_inputs"`
	TotalOutputs int `json:"total_outputs"`
	WebsiteHits  int `json:"website_hits"`
	SentenceHits int `json:"sentence_hits"`

	MemoryUsage MemoryUsage `json:"memory_usage"`

	Logs []string `json:"logs"`
}

type MemoryUsage struct {
	Allocated      uint64 `json:"allocated"`
	TotalAllocated uint64 `json:"total_allocated"`
	System         uint64 `json:"system"`
}

var (
	StartTime     time.Time
	IntakePerHour int
	previousTotal int
	Logs          []string
)

func Start() {
	StartTime = time.Now()

	go intakePerHour()
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

	stats.WriteMode = markov.WriteMode()
	stats.TimeUntilWrite = markov.TimeUntilWrite()
	stats.CurrentCount = markov.CurrentInputs
	stats.CountLimit = markov.WriteInputLimit
	stats.IntakePerHour = IntakePerHour
	stats.Workers = len(markov.CurrentChains())
	stats.PeakIntake = markov.PeakIntake()

	stats.TotalInputs = markov.TotalInputs
	stats.TotalOutputs = markov.TotalOutputs

	stats.MemoryUsage = PrintMemUsage()

	stats.Logs = Logs

	return stats
}

func intakePerHour() {
	for range time.Tick(1 * time.Hour) {
		IntakePerHour = markov.TotalInputs - previousTotal
		previousTotal = markov.TotalInputs
	}
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

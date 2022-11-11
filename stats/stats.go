package stats

import (
	"fmt"
	"log"
	"markov-generator/markov"
	"runtime"
	"time"
)

var (
	StartTime           time.Time
	InputsPerHour       int
	previousIntakeTotal int
	OutputsPerHour      int
	previousOutputTotal int
	Logs                []string
)

func Start() {
	StartTime = time.Now()

	go intakePerHour()
}

func intakePerHour() {
	for range time.Tick(1 * time.Hour) {
		stats := markov.Stats()

		InputsPerHour = stats.SessionInputs - previousIntakeTotal
		previousIntakeTotal = stats.SessionInputs

		OutputsPerHour = stats.SessionOutputs - previousOutputTotal
		previousOutputTotal = stats.SessionOutputs
	}
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
	stats.Markov = markov.Stats()

	stats.InputsPerHour = InputsPerHour
	stats.OutputsPerHour = OutputsPerHour

	stats.MemoryUsage = MemUsage()

	stats.Logs = Logs

	return stats
}

// MemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func MemUsage() (mu MemoryUsage) {
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

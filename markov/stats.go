package markov

import (
	"encoding/json"
	"os"
	"time"
)

type Statistics struct {
	// Start times
	LifetimeStartTime time.Time
	SessionStartTime  time.Time

	// Uptimes
	LifetimeUptime time.Duration
	SessionUptime  time.Duration

	// Inputs
	LifetimeInputs int
	SessionInputs  int

	// Outputs
	LifetimeOutputs int
	SessionOutputs  int

	// Write variables
	WriteMode         string
	InputCurrentCount int
	InputCountLimit   int
	NextWriteTime     time.Time
	TimeUntilWrite    time.Duration

	Workers int

	PeakChainIntake PeakIntakeStruct

	Durations []report
}

type report struct {
	ProcessName string
	ChainName   string
	Duration    time.Duration
}

func updateStats() {
	stats.LifetimeUptime = time.Now().Sub(stats.LifetimeStartTime)
	stats.SessionUptime = time.Now().Sub(stats.SessionStartTime)

	stats.WriteMode = writeMode
	stats.InputCurrentCount = writeInputsCounter
	stats.InputCountLimit = writeInputLimit
	stats.TimeUntilWrite = stats.NextWriteTime.Sub(time.Now())

	stats.Workers = len(CurrentWorkers())
}

func Stats() (statistics Statistics) {
	updateStats()

	return stats
}

func saveStats() {
	Stats()

	statsData, err := json.MarshalIndent(stats, "", " ")
	if err != nil {
		debugLog(err)
	}

	f, err := os.OpenFile("./markov-chains/stats/stats.json", os.O_CREATE, 0666)
	if err != nil {
		debugLog(err)
	}

	_, err = f.Write(statsData)
	defer f.Close()

	if err != nil {
		//debugLog("wrote unsuccessfully to", "./markov-chains/stats/stats.json")
		debugLog(err)
	} else {
		//debugLog("wrote successfully to", "./markov-chains/stats/stats.json")
	}
}

func loadStats() {
	f, err := os.OpenFile("./markov-chains/stats/stats.json", os.O_CREATE, 0666)
	if err != nil {
		debugLog("Failed reading stats:", err)
	}
	defer f.Close()

	fS, _ := f.Stat()
	if fS.Size() == 0 {
		stats.LifetimeStartTime = time.Now()
		stats.SessionStartTime = time.Now()
		return
	}

	err = json.NewDecoder(f).Decode(&stats)
	if err != nil {
		debugLog("Error when unmarshalling stats:", "\n", err)
	}

	stats.SessionStartTime = time.Now()
	stats.SessionInputs = 0
	stats.SessionOutputs = 0
}

func track(process string, chain string) (string, string, time.Time) {
	return process, chain, time.Now()
}

func duration(process string, chain string, start time.Time) {
	duration := time.Since(start).Round(1 * time.Second)
	debugLog(process + ": " + duration.String())

	stats.Durations = append(stats.Durations, report{ProcessName: process, ChainName: chain, Duration: duration})
}

func ReportDurations() []report {
	return stats.Durations
}

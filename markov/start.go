package markov

import (
	"time"
)

var (
	writeMode       string
	writeInterval   int
	intervalUnit    string
	WriteCountLimit int
	startKey        string
	endKey          string
	Debug           bool

	nextWriteTime   time.Time
	chainPeakIntake struct {
		Chain  string
		Amount int
		Time   time.Time
	}
)

// Start initializes the Markov  package.
//
// Takes:
//		StartInstructions {
//			WriteMode     string
//			WriteInterval int
//			IntervalUnit  string
//			WriteLimit    int
//			StartKey      string
//			EndKey        string
//			Debug         bool
// 		}
func Start(instructions StartInstructions) {
	createChainsFolder()

	writeMode = instructions.WriteMode
	writeInterval = instructions.WriteTicker
	intervalUnit = instructions.TickerUnit
	WriteCountLimit = instructions.WriteLimit
	startKey = instructions.StartKey
	endKey = instructions.EndKey
	Debug = instructions.Debug

	toWorker = make(chan input)

	startWorkers()

	go distributor()

	if writeMode == "ticker" {
		go writeTicker()
	}
}

func writeCounter() {
	if writeMode == "counter" {
		CurrentCount += 1
		if CurrentCount > WriteCountLimit {
			go writeLoop()
			CurrentCount = 0
		}
	}
}

func writeTicker() {
	var unit time.Duration

	switch intervalUnit {
	default:
		unit = time.Minute
	case "seconds":
		unit = time.Second
	case "minutes":
		unit = time.Minute
	case "hours":
		unit = time.Hour
	}

	nextWriteTime = time.Now().Add(time.Duration(writeInterval) * unit)
	debugLog("write ticker started")
	for range time.Tick(time.Duration(writeInterval) * unit) {
		nextWriteTime = time.Now().Add(time.Duration(writeInterval) * unit)
		writeLoop()
	}
}

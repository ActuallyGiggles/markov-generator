package markov

import (
	"time"
)

var (
	TotalInputs     int
	CurrentInputs   int
	nextWriteTime   time.Time
	peakChainIntake PeakIntakeStruct
)

func writeCounter() {
	TotalInputs += 1
	if writeMode == "counter" {
		CurrentInputs += 1
		if CurrentInputs > WriteInputLimit {
			go writeLoop()
			CurrentInputs = 0
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
	for range time.Tick(time.Duration(writeInterval) * unit) {
		nextWriteTime = time.Now().Add(time.Duration(writeInterval) * unit)
		go writeLoop()
	}
}

func writeLoop() {
	//writing := 0
	for _, w := range workerMap {
		if w.Intake == 0 {
			continue
		}

		if w.Intake > peakChainIntake.Amount {
			peakChainIntake.Chain = w.Name
			peakChainIntake.Amount = w.Intake
			peakChainIntake.Time = time.Now()
		}

		// if writing >= len(workerMap)/2 {
		// 	w.writeToChain()
		// 	writing = 0
		// 	continue
		// }

		w.writeToFile()

		//writing += 1
	}
}

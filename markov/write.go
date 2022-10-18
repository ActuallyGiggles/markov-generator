package markov

import (
	"time"
)

var (
	CurrentCount    int
	nextWriteTime   time.Time
	chainPeakIntake struct {
		Chain  string
		Amount int
		Time   time.Time
	}
)

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

		if w.Intake > chainPeakIntake.Amount {
			chainPeakIntake.Chain = w.Name
			chainPeakIntake.Amount = w.Intake
			chainPeakIntake.Time = time.Now()
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

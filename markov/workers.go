package markov

import (
	"sync"
	"time"
)

var (
	workerMap    = make(map[string]*worker)
	workerMapMx  sync.Mutex
	toWorker     chan input
	CurrentCount int
)

func startWorkers() {
	chains := chains()
	for _, name := range chains {
		newWorker(name)
	}
}

func newWorker(name string) *worker {
	w := &worker{
		ChainResponsibleFor: name,
		ChainToWrite:        make(map[string]map[string]map[string]int),
	}

	workerMapMx.Lock()
	workerMap[name] = w
	workerMapMx.Unlock()

	w.Status = "Ready"
	w.LastModified = now()
	debugLog("Created worker:", name)

	return w
}

func distributor() {
	for in := range toWorker {
		//if in.Chain != "hasanabi" {
		//	continue
		//}

		if in.Content == "" {
			debugLog("empty muthafuckin uhhhhhh content yo", in.Chain, in.Content)
			continue
		}

		//debugLog("distributor locks", in.Chain)
		workerMapMx.Lock()
		worker, ok := workerMap[in.Chain]
		workerMapMx.Unlock()
		//debugLog("distributor unlocks", in.Chain)

		if worker.Status == "Ready" {
			if ok {
				go worker.addToQueue(in.Chain, in.Content)
			} else {
				worker = newWorker(in.Chain)
				go worker.addToQueue(in.Chain, in.Content)
			}
		}
	}
}

func (w *worker) addToQueue(chain string, content string) {
	w.ChainToWriteMx.Lock()
	//debugLog("addToQueue locks", w.ChainResponsibleFor)

	contentToChain(&w.ChainToWrite, chain, content)
	w.Intake += 1

	w.ChainToWriteMx.Unlock()
	//debugLog("addToQueue unlocks", w.ChainResponsibleFor)

	writeCounter()
}

func writeLoop() {
	debugLog("write loop started")
	defer debugLog("write loop done")
	//writing := 0
	for _, w := range workerMap {
		if w.Intake == 0 || w.Status == "Writing" {
			continue
		}

		if w.Intake > chainPeakIntake.Amount {
			chainPeakIntake.Chain = w.ChainResponsibleFor
			chainPeakIntake.Amount = w.Intake
			chainPeakIntake.Time = time.Now()
		}

		// if writing >= len(workerMap)/2 {
		// 	w.writeToChain()
		// 	writing = 0
		// 	continue
		// }

		w.writeToChain()

		//writing += 1
	}
}

func (w *worker) writeToChain() {
	defer duration(track(w.ChainResponsibleFor))

	w.ChainToWriteMx.Lock()
	debugLog("writeToChain locks", w.ChainResponsibleFor)

	w.Status = "Writing"
	w.LastModified = now()
	path := "./markov/chains/" + w.ChainResponsibleFor + ".json"

	existingChain, chainExists := jsonToChain(path)
	if !chainExists {
		existingChain = make(map[string]map[string]map[string]int)
		debugLog("CHAIN CREATED FOR", w.ChainResponsibleFor)
	}

	for currentParent, currentParentValue := range w.ChainToWrite {
		if _, ok := existingChain[currentParent]; !ok {
			existingChain[currentParent] = make(map[string]map[string]int)
		}
		for currentList, currentListValue := range currentParentValue {
			if _, ok := existingChain[currentParent][currentList]; !ok {
				existingChain[currentParent][currentList] = make(map[string]int)
			}
			for currentChild, currentTimesUsed := range currentListValue {
				existingChain[currentParent][currentList][currentChild] += currentTimesUsed
			}
		}
	}

	chainToJson(existingChain, path)

	w.ChainToWrite = make(map[string]map[string]map[string]int)
	w.Intake = 0
	w.Status = "Ready"
	w.LastModified = now()

	w.ChainToWriteMx.Unlock()
	debugLog("writeToChain unlocks", w.ChainResponsibleFor)
}

// WorkersStats returns a slice of type WorkerStats
func WorkersStats() (slice []WorkerStats) {
	workerMapMx.Lock()
	for _, w := range workerMap {
		e := WorkerStats{
			ChainResponsibleFor: w.ChainResponsibleFor,
			Intake:              w.Intake,
			Status:              w.Status,
			LastModified:        w.LastModified,
		}
		slice = append(slice, e)
	}
	workerMapMx.Unlock()
	return slice
}

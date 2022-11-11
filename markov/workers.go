package markov

import (
	"sync"
)

var (
	workerMap   = make(map[string]*worker)
	workerMapMx sync.Mutex
)

func startWorkers() {
	chains := chains()
	for _, name := range chains {
		newWorker(name)
	}
}

func newWorker(name string) *worker {
	var w *worker
	w = &worker{
		Name:  name,
		Chain: chain{},
	}

	workerMapMx.Lock()
	workerMap[name] = w
	workerMapMx.Unlock()

	return w
}

// WorkersStats returns a slice of type WorkerStats.
func WorkersStats() (slice []WorkerStats) {
	workerMapMx.Lock()
	for _, w := range workerMap {
		e := WorkerStats{
			ChainResponsibleFor: w.Name,
			Intake:              w.Intake,
		}
		slice = append(slice, e)
	}
	workerMapMx.Unlock()
	return slice
}

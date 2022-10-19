package markov

var (
	writeMode       string
	writeInterval   int
	intervalUnit    string
	WriteCountLimit int
	startKey        string
	endKey          string
	debug           bool
)

// Start starts markov based on instructions provided.
func Start(sI StartInstructions) error {
	writeMode = sI.WriteMode
	writeInterval = sI.WriteInterval
	intervalUnit = sI.IntervalUnit
	WriteCountLimit = sI.WriteLimit
	startKey = sI.StartKey
	endKey = sI.EndKey
	debug = sI.Debug

	createChainsFolder()

	startWorkers()

	if writeMode == "interval" {
		go writeTicker()
	}

	return nil
}

func startWorkers() {
	chains := chains()
	for _, name := range chains {
		newWorker(name)
	}
}

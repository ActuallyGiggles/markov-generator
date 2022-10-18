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

func Start(sI StartInstructions) error {
	writeMode = sI.WriteMode
	writeInterval = sI.WriteTicker
	intervalUnit = sI.TickerUnit
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

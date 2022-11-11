package markov

var (
	writeMode       string
	writeInterval   int
	intervalUnit    string
	writeInputLimit int
	startKey        string
	endKey          string
	debug           bool

	stats Statistics
)

// Start starts markov based on instructions provided.
func Start(sI StartInstructions) error {
	writeMode = sI.WriteMode
	writeInterval = sI.WriteInterval
	intervalUnit = sI.IntervalUnit
	writeInputLimit = sI.WriteLimit
	startKey = sI.StartKey
	endKey = sI.EndKey
	debug = sI.Debug

	createFolders()

	loadStats()

	startWorkers()

	if writeMode == "interval" {
		go writeTicker()
	}

	return nil
}

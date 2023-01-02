package markov

var (
	writeMode       string
	writeInterval   int
	intervalUnit    string
	writeInputLimit int
	startKey        string
	endKey          string
	shouldZip       bool
	debug           bool

	zipping bool
	writing bool

	durations chan string

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

	durations = sI.ReportDurations

	shouldZip = sI.ShouldZip
	debug = sI.Debug

	createFolders()

	loadStats()

	startWorkers()

	if writeMode == "interval" {
		go writeTicker()
	}

	if shouldZip {
		go zipTicker()
	}

	return nil
}

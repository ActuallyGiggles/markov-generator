package markov

import "sync"

// StartInstructions are the instructions on how Markov should be started
//
//		WriteMode: The method to use to initiate a write cycle.
//			"counter"
//			"ticker"
// 		Workers: How many workers to create
// 		WriteTicker: How often to write to chains (in minutes)
//		TickerUnit: What unit to use for WriteTicker (default minutes)
//			"seconds"
//			"minutes"
//			"hours"
//		WriteLimit: On what amount of inputs should a write cycle start (applies to all chains)
//		StartKey: A string that is used to signify the natural beginning to the content
// 		EndKey: A string that is used to signify the natural end to the content
//
// (StartKey and EndKey should be strings that are unlikely to be natural content)
type StartInstructions struct {
	WriteMode   string
	WriteTicker int
	TickerUnit  string
	WriteLimit  int
	StartKey    string
	EndKey      string
	Debug       bool
}

// WorkerStats contains the current statistics on a worker
//
// 		ID: ID of worker
//		Intake: How many inputs this worker has gotten (clears after every Instructions.WriteTicker)
//		Status: Current status of worker
//			"Ready": Accepting content into queue
//			"Writing": Writing queue into chains
//		LastModified: The last time the status was updated
type WorkerStats struct {
	ChainResponsibleFor string
	Intake              int
	Status              string
	LastModified        string
}

// OutputInstructions are the instructions to give Markov when asking for an output
//
// 		Method: Which method to use when constructing the output
//			"LikelyBeginning": Build from a likely beginning
//			"TargetedBeginning": Build from a targeted beginning
//			"LikelyEnding": Build from a likely Ending (to be added)
//			"TargetedEnding": Build from a targeted ending (to be added)
//			"TargetedMiddle": Build from a targeted middle (to be added)
//		Chain: Which chain to use
// 		Target: Optional target (leave blank if N/A)
type OutputInstructions struct {
	Method string
	Chain  string
	Target string
}

type input struct {
	Chain   string
	Content string
}

type result struct {
	Output  string
	Problem string
}

type worker struct {
	ChainResponsibleFor string
	ChainToWrite        map[string]map[string]map[string]int
	ChainToWriteMx      sync.Mutex
	Intake              int
	Status              string
	LastModified        string
}

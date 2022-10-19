package markov

import "sync"

// StartInstructions details intructions to start markov.
//
// 	WriteMode: What triggers a writing cycle. ("counter" or "interval")
// 	WriteInterval: How often to trigger a write cycle.
// 	IntervalUnit: What unit to use for the WriteInterval.
// 	WriteLimit: To trigger a write cycle after x amount of ins, or entries.
// 	StartKey: What string can be used to mark the beginning of a message. (E.g. !-)
// 	EndKey: What string can be used to mark the end of a message. (E.g. -!)
// 	Debug: Print logs of stuffs.
type StartInstructions struct {
	WriteMode     string
	WriteInterval int
	IntervalUnit  string
	WriteLimit    int
	StartKey      string
	EndKey        string
	Debug         bool
}

// OutputInstructions details instructions on how to make an output.
//
// 	Chain: What chain to use.
// 	Method: What method to use.
// 		"LikelyBeginning": Start with a likely beginning word.
//		"TargetedBeginning": Start with a specific beginning word.
// 		"TargetedMiddle": Generate a message with a specific middle word. (yet to implement)
// 		"LikelyEnd": End with a likely ending word. (yet to implement)
//		"TargetedEnd": End with a specific ending word. (yet to implement)
type OutputInstructions struct {
	Chain  string
	Method string
	Target string
}

type worker struct {
	Name    string
	Chain   chain
	ChainMx sync.Mutex
	Intake  int
}

type chain struct {
	Parents []parent `json:"parents"`
}

type parent struct {
	Word     string `json:"word"`
	Next     []word `json:"next"`
	Previous []word `json:"previous"`
}

type word struct {
	Word  string `json:"word"`
	Value int    `json:"value"`
}

type input struct {
	Name    string
	Content string
}

type wRand struct {
	Word  string
	Value int
}

// WorkerStats contains the name of the chain the worker is responsible for and the intake amount in that worker.
type WorkerStats struct {
	ChainResponsibleFor string
	Intake              int
}

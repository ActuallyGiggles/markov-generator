package markov

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// StartInstructions details intructions to start markov.
//
// 	WriteMode: What triggers a writing cycle. ("counter" or "interval")
// 	WriteInterval: How often to trigger a write cycle.
// 	IntervalUnit: What unit to use for the WriteInterval.
// 	WriteLimit: To trigger a write cycle after x amount of ins, or entries.
// 	StartKey: What string can be used to mark the beginning of a message. (E.g. !-)
// 	EndKey: What string can be used to mark the end of a message. (E.g. -!)
//  ReportDurations: If you would like the bot to report the duration of writing cycles or zip cycles, provide a channel. Otherwise, leave as nil.
//  Zip: Whether or not to zip the markov-chains folder every 24 hours.
// 	Debug: Print logs of stuffs.
type StartInstructions struct {
	WriteMode string

	WriteInterval int
	IntervalUnit  string

	WriteLimit int

	StartKey string
	EndKey   string

	ReportDurations chan string

	ShouldZip bool

	Debug bool
}

// OutputInstructions details instructions on how to make an output.
//
// 	Chain: What chain to use.
// 	Method: What method to use.
// 		"LikelyBeginning": Start with a likely beginning word.
//		"TargetedBeginning": Start with a specific beginning word.
// 		"TargetedMiddle": Generate a message with a specific middle word. (yet to implement)
//		"TargetedEnding": End with a specific ending word.
// 		"LikelyEnding": End with a likely ending word.
type OutputInstructions struct {
	Chain  string
	Method string
	Target string
}

type worker struct {
	Name    string
	Chain   chain
	ChainMx sync.RWMutex
	Intake  int
}

type chain struct {
	Parents []parent
}

type parent struct {
	Word         string
	Children     []child
	Grandparents []grandparent
}

type child struct {
	Word  string
	Value int
}

type grandparent struct {
	Word  string
	Value int
}

type input struct {
	Name    string
	Content string
}

// WorkerStats contains the name of the chain the worker is responsible for and the intake amount in that worker.
type WorkerStats struct {
	ChainResponsibleFor string
	Intake              int
}

type PeakIntakeStruct struct {
	Chain  string    `json:"chain"`
	Amount int       `json:"amount"`
	Time   time.Time `json:"time"`
}

// A Choice contains a generic item and a weight controlling the frequency with
// which it will be selected.
type Choice struct {
	Weight int
	Word   string
}

type encode struct {
	Encoder        *json.Encoder
	File           *os.File
	ContinuedEntry bool
}

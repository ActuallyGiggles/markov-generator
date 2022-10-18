package markov

import "sync"

type StartInstructions struct {
	WriteMode   string
	WriteTicker int
	TickerUnit  string
	WriteLimit  int
	StartKey    string
	EndKey      string
	Debug       bool
}

type OutputInstructions struct {
	Chain  string
	Method string
	Target string
}

type worker struct {
	Name         string
	Chain        chain
	ChainMx      sync.Mutex
	Intake       int
	Status       string
	LastModified string
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

type WorkerStats struct {
	ChainResponsibleFor string
	Intake              int
	Status              string
	LastModified        string
}

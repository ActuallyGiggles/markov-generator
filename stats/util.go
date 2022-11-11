package stats

import (
	"markov-generator/markov"
)

type Stats struct {
	Markov         markov.Statistics
	InputsPerHour  int
	OutputsPerHour int

	MemoryUsage MemoryUsage `json:"memory_usage"`

	WebsiteHits  int
	SentenceHits int

	Logs []string `json:"logs"`
}

type MemoryUsage struct {
	Allocated      uint64 `json:"allocated"`
	TotalAllocated uint64 `json:"total_allocated"`
	System         uint64 `json:"system"`
}

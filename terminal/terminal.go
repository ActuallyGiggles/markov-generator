package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/ActuallyGiggles/go-markov"
)

type Terminal struct {
	Markov      string
	StartTime   time.Time
	RunningTime time.Duration
	Live        string
	Emotes      string
	Workers     []WorkerForTerminal
}

type WorkerForTerminal struct {
	ID     int
	Intake int
	Status string
	Time   string
}

var T Terminal

func UpdateTerminal(mode string) {
	dt := time.Now()
	now := dt.Format("15:04:05")

	switch mode {
	case "init":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		T.Markov = "Active"
		fmt.Printf("\tMarkov: %s", T.Markov)
		T.StartTime = time.Now()
		go refreshTerminal()
	case "live":
		T.Live = now
	case "emotes":
		T.Emotes = now
	}
}

func refreshTerminal() {
	for range time.Tick(5 * time.Second) {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		T.RunningTime = time.Now().Sub(T.StartTime)
		fmt.Printf("\tMarkov:       %s\n\tRunning time: %s\n\tLive:         %s\n\tEmotes:       %s\n\n", T.Markov, T.RunningTime, T.Live, T.Emotes)

		workers := markov.WorkersStats()
		sort.Slice(workers, func(i, j int) bool {
			return workers[i].ID < workers[j].ID
		})
		fmt.Println("\tNext Write in:", markov.TimeUntilWrite())
		fmt.Println()
		for _, worker := range workers {
			fmt.Printf("\tWorker %02d\t%04d\t%s", worker.ID, worker.Intake, worker.Status)
			fmt.Println()
		}
	}
}

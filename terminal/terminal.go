package terminal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"MarkovGenerator/markov"
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
	tn := time.Now()
	now := tn.Format("15:04:05")

	switch mode {
	case "init":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		T.Markov = "Active"
		log.Printf("\tMarkov: %s", T.Markov)
		T.StartTime = tn
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
		fmt.Println("\tMarkov:       ", T.Markov)
		fmt.Println("\tRunning time: ", T.RunningTime)
		fmt.Println("\tLive:         ", T.Live)
		fmt.Println("\tEmotes:       ", T.Emotes)

		workers := markov.WorkersStats()
		sort.Slice(workers, func(i, j int) bool {
			return workers[i].ID < workers[j].ID
		})

		fmt.Println()
		pi := markov.PeakIntake()
		fmt.Println("\tPeak intake:  ", pi.Amount, pi.Time.Format("15:04:05"))
		fmt.Println("\tNext write in:", markov.TimeUntilWrite())
		for _, worker := range workers {
			fmt.Printf("\tWorker %02d\t%04d\t%s\n", worker.ID, worker.Intake, worker.Status)
		}
	}
}

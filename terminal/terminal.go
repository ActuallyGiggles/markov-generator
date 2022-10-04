package terminal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"markov-generator/markov"
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
			return workers[i].ChainResponsibleFor < workers[j].ChainResponsibleFor
		})

		fmt.Println()
		pi := markov.PeakIntake()
		fmt.Println("\tPeak intake:  ", pi.Amount, pi.Time.Format("15:04:05"))
		fmt.Println("\tNext write in:", markov.TimeUntilWrite())

		for i := 0; i < len(workers); i += 3 {
			worker := workers[i]
			fmt.Printf("\t  %-20s\t%04d\t%-5s | ", worker.ChainResponsibleFor, worker.Intake, worker.Status)

			if exists := doesSliceContainIndex(workers, i+1); exists {
				worker2 := workers[i+1]
				fmt.Printf("%-20s\t%04d\t%-5s | ", worker2.ChainResponsibleFor, worker2.Intake, worker2.Status)
			}

			if exists := doesSliceContainIndex(workers, i+2); exists {
				worker3 := workers[i+2]
				fmt.Printf("%-20s\t%04d\t%-5s", worker3.ChainResponsibleFor, worker3.Intake, worker3.Status)
			}

			fmt.Println()
		}
	}
}

func doesSliceContainIndex(slice []markov.WorkerStats, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

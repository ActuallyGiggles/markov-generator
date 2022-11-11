package markov

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var writeInputsCounter int

func writeCounter() {
	if writeMode == "counter" {
		writeInputsCounter++
		if writeInputsCounter > writeInputLimit {
			go writeLoop()
		}
	}
}

func writeTicker() {
	var unit time.Duration

	switch intervalUnit {
	default:
		unit = time.Minute
	case "seconds":
		unit = time.Second
	case "minutes":
		unit = time.Minute
	case "hours":
		unit = time.Hour
	}

	stats.NextWriteTime = time.Now().Add(time.Duration(writeInterval) * unit)
	for range time.Tick(time.Duration(writeInterval) * unit) {
		stats.NextWriteTime = time.Now().Add(time.Duration(writeInterval) * unit)
		go writeLoop()
	}
}

func writeLoop() {
	for _, w := range workerMap {
		if w.Intake == 0 {
			continue
		}

		// Find new peak intake chain
		if w.Intake > stats.PeakChainIntake.Amount {
			stats.PeakChainIntake.Chain = w.Name
			stats.PeakChainIntake.Amount = w.Intake
			stats.PeakChainIntake.Time = time.Now()
		}

		w.writeToFile()
	}

	writeInputsCounter = 0
	saveStats()
}

func (w *worker) writeToFile() {
	defer duration(track("INPUT: " + w.Name))

	w.ChainMx.Lock()
	defer w.ChainMx.Unlock()

	// Specify updated list of parents
	var updatedChain []parent

	// Open existing chain file
	f, err := os.OpenFile("./markov-chains/"+w.Name+".json", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			chainData, _ := json.MarshalIndent(w.Chain.Parents, "", "    ")
			f.Write(chainData)
			f.Close()
			return
		}

		// For every new item in the existing chain
		for dec.More() {
			var existingParent parent

			err = dec.Decode(&existingParent)
			if err != nil {
				panic(err)
			}

			parentMatch := false
			// Find newParent in existingParents
			for nPIndex := 0; nPIndex < len(w.Chain.Parents); nPIndex++ {
				newParent := &w.Chain.Parents[nPIndex]
				if newParent.Word == existingParent.Word {
					parentMatch = true

					uParent := parent{
						Word: newParent.Word,
					}

					// combine values and set into updatedChain
					for eCIndex := 0; eCIndex < len(existingParent.Children); eCIndex++ {
						existingChild := &existingParent.Children[eCIndex]
						childMatch := false
						for nCIndex := 0; nCIndex < len(newParent.Children); nCIndex++ {
							newChild := &newParent.Children[nCIndex]
							if newChild.Word == existingChild.Word {
								childMatch = true

								uParent.Children = append(uParent.Children, child{
									Word:  newChild.Word,
									Value: newChild.Value + existingChild.Value,
								})

								newParent.Children = removeChild(newParent.Children, nCIndex)
								break
							}
						}

						if !childMatch {
							uParent.Children = append(uParent.Children, *existingChild)
						}
					}

					for _, newChild := range newParent.Children {
						uParent.Children = append(uParent.Children, newChild)
					}

					updatedChain = append(updatedChain, uParent)

					w.Chain.Parents = removeParent(w.Chain.Parents, nPIndex)
					break
				}
			}

			if !parentMatch {
				updatedChain = append(updatedChain, existingParent)
			}
		}

		for _, nParent := range w.Chain.Parents {
			updatedChain = append(updatedChain, nParent)
		}
	}

	// Close the chain file
	f.Close()

	// Open file again
	f, err = os.OpenFile("./markov-chains/"+w.Name+".json", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	// Indent and write
	chainData, err := json.MarshalIndent(updatedChain, "", "    ")
	if err != nil {
		panic(err)
	}
	_, err = f.Write(chainData)
	if err != nil {
		fmt.Println(err)
	}
	f.Close()
}

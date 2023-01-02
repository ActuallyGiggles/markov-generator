package markov

import (
	"encoding/json"
	"os"
	"time"
)

var writeInputsCounter int

func writeCounter() {
	if writeMode != "counter" {
		return
	}

	writeInputsCounter++

	if writeInputsCounter < writeInputLimit {
		return
	}

	if zipping {
		return
	}

	go writeLoop()
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
	if writing {
		return
	}

zippingLoop:
	debugLog("write ticker went off")
	if zipping {
		debugLog("cannot write because currently zipping -- trying again in 30s")
		time.Sleep(30 * time.Second)
		goto zippingLoop
	}

	writing = true

	debugLog("write loop started")

	defer duration(track("write duration", ""))

	for _, w := range workerMap {
		tP, tC, tT := track("write", w.Name)

		w.ChainMx.Lock()

		if w.Intake == 0 {
			w.ChainMx.Unlock()
			continue
		}

		// Find new peak intake chain
		if w.Intake > stats.PeakChainIntake.Amount {
			stats.PeakChainIntake.Chain = w.Name
			stats.PeakChainIntake.Amount = w.Intake
			stats.PeakChainIntake.Time = time.Now()
		}

		w.writeHead()
		w.writeTail()
		w.writeBody()

		w.Chain.Parents = nil
		w.Intake = 0

		w.ChainMx.Unlock()

		duration(tP, tC, tT)
	}

	writeInputsCounter = 0
	saveStats()

	writing = false
}

func (w *worker) writeHead() {
	defaultPath := "./markov-chains/" + w.Name + "_head.json"
	newPath := "./markov-chains/" + w.Name + "_head_new.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			chainData, _ := json.MarshalIndent(w.Chain.Parents[0].Children, "", "    ")
			w.Chain.removeParent(0)
			f.Write(chainData)
			f.Close()
			return
		} else {
			fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}

			var enc encode

			if err = StartEncoder(&enc, fN); err != nil {
				panic(err)
			}

			for i, parent := range *&w.Chain.Parents {
				if parent.Word == startKey {

					for dec.More() {
						var existingChild child

						err := dec.Decode(&existingChild)
						if err != nil {
							panic(err)
						}

						childMatch := false

						for j, newChild := range *&parent.Children {

							if newChild.Word == existingChild.Word {
								childMatch = true

								enc.AddEntry(child{
									Word:  newChild.Word,
									Value: newChild.Value + existingChild.Value,
								})

								parent.removeChild(j)

								continue
							}
						}

						if !childMatch {
							enc.AddEntry(existingChild)
						}
					}

					for _, c := range *&parent.Children {
						enc.AddEntry(c)
					}

					w.Chain.removeParent(i)
				}
			}

			enc.CloseEncoder()
			fN.Close()
		}
	}

	f.Close()

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func (w *worker) writeTail() {
	defaultPath := "./markov-chains/" + w.Name + "_tail.json"
	newPath := "./markov-chains/" + w.Name + "_tail_new.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			chainData, _ := json.MarshalIndent(w.Chain.Parents[0].Grandparents, "", "    ")
			w.Chain.removeParent(0)
			f.Write(chainData)
			f.Close()
			return
		} else {
			fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}

			var enc encode

			if err = StartEncoder(&enc, fN); err != nil {
				panic(err)
			}

			for i, parent := range *&w.Chain.Parents {
				if parent.Word == endKey {

					for dec.More() {
						var existingGrandparent grandparent

						err := dec.Decode(&existingGrandparent)
						if err != nil {
							panic(err)
						}

						grandparentMatch := false

						for j, newGrandparent := range *&parent.Grandparents {

							if newGrandparent.Word == existingGrandparent.Word {
								grandparentMatch = true

								enc.AddEntry(child{
									Word:  newGrandparent.Word,
									Value: newGrandparent.Value + existingGrandparent.Value,
								})

								parent.removeGrandparent(j)

								continue
							}
						}

						if !grandparentMatch {
							enc.AddEntry(existingGrandparent)
						}
					}

					for _, c := range *&parent.Grandparents {
						enc.AddEntry(c)
					}

					w.Chain.removeParent(i)
				}
			}

			enc.CloseEncoder()
			fN.Close()
		}
	}

	f.Close()

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func (w *worker) writeBody() {
	defaultPath := "./markov-chains/" + w.Name + "_body.json"
	newPath := "./markov-chains/" + w.Name + "_body_new.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
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
		} else {
			fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}

			var enc encode

			StartEncoder(&enc, fN)

			// For every new item in the existing chain
			for dec.More() {
				var existingParent parent

				err = dec.Decode(&existingParent)
				if err != nil {
					panic(err)
				}

				parentMatch := false
				// Find newParent in existingParents
				for nPIndex, newParent := range *&w.Chain.Parents {

					if newParent.Word == existingParent.Word {
						parentMatch = true

						uParent := parent{
							Word: newParent.Word,
						}

						// Do for child
						// combine values and set into updatedChain
						for _, existingChild := range *&existingParent.Children {
							childMatch := false

							for nCIndex, newChild := range *&newParent.Children {

								if newChild.Word == existingChild.Word {
									childMatch = true

									uParent.Children = append(uParent.Children, child{
										Word:  newChild.Word,
										Value: newChild.Value + existingChild.Value,
									})

									newParent.removeChild(nCIndex)
									break
								}
							}

							if !childMatch {
								uParent.Children = append(uParent.Children, existingChild)
							}
						}

						for _, newChild := range newParent.Children {
							uParent.Children = append(uParent.Children, newChild)
						}

						// Do for grandparent
						// combine values and set into updatedChain
						for _, existingGrandparent := range *&existingParent.Grandparents {
							grandparentMatch := false

							for nPIndex, newGrandparent := range *&newParent.Grandparents {

								if newGrandparent.Word == existingGrandparent.Word {
									grandparentMatch = true

									uParent.Grandparents = append(uParent.Grandparents, grandparent{
										Word:  newGrandparent.Word,
										Value: newGrandparent.Value + existingGrandparent.Value,
									})

									newParent.removeGrandparent(nPIndex)
									break
								}
							}

							if !grandparentMatch {
								uParent.Grandparents = append(uParent.Grandparents, existingGrandparent)
							}
						}

						for _, newGrandparent := range newParent.Grandparents {
							uParent.Grandparents = append(uParent.Grandparents, newGrandparent)
						}

						enc.AddEntry(uParent)

						w.Chain.removeParent(nPIndex)
						break
					}
				}

				if !parentMatch {
					enc.AddEntry(existingParent)
				}
			}

			for _, nParent := range w.Chain.Parents {
				enc.AddEntry(nParent)
			}

			enc.CloseEncoder()
			fN.Close()
		}
	}

	// Close the chain file
	f.Close()

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

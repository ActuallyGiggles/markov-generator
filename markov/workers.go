package markov

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	workerMap   = make(map[string]*worker)
	workerMapMx sync.Mutex
)

func newWorker(name string) *worker {
	var w *worker
	w = &worker{
		Name:  name,
		Chain: chain{},
	}

	workerMapMx.Lock()
	workerMap[name] = w
	workerMapMx.Unlock()

	return w
}

func (w *worker) addInput(content string) {
	w.ChainMx.Lock()
	defer w.ChainMx.Unlock()

	contentToChain(&w.Chain, w.Name, content)
	w.Intake += 1
}

func (w *worker) writeToFile() {
	defer duration(track(w.Name))

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

	// eC, err := jsonToChain(w.Name)
	// if err != nil {
	// 	chainToJson(w.Chain, w.Name)
	// 	w.Intake = 0
	// 	w.Chain = chain{}
	// } else {
	// 	var chainToWrite chain

	// 	for _, nParent := range w.Chain.Parents {
	// 		parentMatch := false
	// 		for eParentIndex, eParent := range eC.Parents {
	// 			if eParent.Word == nParent.Word {
	// 				parentMatch = true

	// 				uParent := parent{
	// 					Word: eParent.Word,
	// 				}

	// 				for _, nChild := range nParent.Children {
	// 					childMatch := false
	// 					for eChildIndex, eChild := range eParent.Children {
	// 						if eChild.Word == nChild.Word {
	// 							childMatch = true

	// 							uParent.Children = append(uParent.Children, child{
	// 								Word:  eChild.Word,
	// 								Value: eChild.Value + nChild.Value,
	// 							})

	// 							eParent.Children = removeChild(eParent.Children, eChildIndex)
	// 						}
	// 					}
	// 					if !childMatch {
	// 						uParent.Children = append(uParent.Children, nChild)
	// 					}
	// 				}

	// 				for _, eChild := range eParent.Children {
	// 					uParent.Children = append(uParent.Children, eChild)
	// 				}

	// 				chainToWrite.Parents = append(chainToWrite.Parents, uParent)
	// 				eC.Parents = removeParent(eC.Parents, eParentIndex)
	// 			}
	// 		}
	// 		if !parentMatch {
	// 			chainToWrite.Parents = append(chainToWrite.Parents, nParent)
	// 		}
	// 	}

	// 	for _, eParent := range eC.Parents {
	// 		chainToWrite.Parents = append(chainToWrite.Parents, eParent)
	// 	}

	// 	chainToJson(chainToWrite, w.Name)
	// 	w.Intake = 0
	// 	w.Chain = chain{}
	// }
}

// WorkersStats returns a slice of type WorkerStats.
func WorkersStats() (slice []WorkerStats) {
	workerMapMx.Lock()
	for _, w := range workerMap {
		e := WorkerStats{
			ChainResponsibleFor: w.Name,
			Intake:              w.Intake,
		}
		slice = append(slice, e)
	}
	workerMapMx.Unlock()
	return slice
}

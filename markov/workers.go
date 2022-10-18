package markov

import (
	"fmt"
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

	// for _, parent := range w.Chain.Parents {
	// 	for _, child := range parent.Next {
	// 		fmt.Println(parent.Word, "->", child.Word)
	// 	}
	// 	for _, grandparent := range parent.Previous {
	// 		fmt.Println(parent.Word, "<-", grandparent.Word)
	// 	}
	// }
}

func (w *worker) writeToFile() {
	defer duration(track(w.Name))

	w.ChainMx.Lock()
	defer w.ChainMx.Unlock()

	eC, err := jsonToChain(w.Name)
	if err != nil {
		fmt.Println(err)
		chainToJson(w.Chain, w.Name)
		w.Intake = 0
		w.Chain = chain{}
	} else {
		var chainToWrite chain

		for _, nParent := range w.Chain.Parents {
			parentMatch := false
			for eParentIndex, eParent := range eC.Parents {
				if eParent.Word == nParent.Word {
					parentMatch = true

					uParent := parent{
						Word: eParent.Word,
					}

					for _, nChild := range nParent.Next {
						childMatch := false
						for eChildIndex, eChild := range eParent.Next {
							if eChild.Word == nChild.Word {
								childMatch = true

								uParent.Next = append(uParent.Next, word{
									Word:  eChild.Word,
									Value: eChild.Value + nChild.Value,
								})

								eParent.Next = removeCorGP(eParent.Next, eChildIndex)
							}
						}
						if !childMatch {
							uParent.Next = append(uParent.Next, nChild)
						}
					}

					for _, eChild := range eParent.Next {
						uParent.Next = append(uParent.Next, eChild)
					}

					for _, nGrandparent := range nParent.Previous {
						GrandparentMatch := false
						for eGrandparentIndex, eGrandparent := range eParent.Previous {
							if eGrandparent.Word == nGrandparent.Word {
								GrandparentMatch = true

								uParent.Previous = append(uParent.Previous, word{
									Word:  eGrandparent.Word,
									Value: eGrandparent.Value + nGrandparent.Value,
								})

								eParent.Previous = removeCorGP(eParent.Previous, eGrandparentIndex)
							}
						}
						if !GrandparentMatch {
							uParent.Previous = append(uParent.Previous, nGrandparent)
						}
					}

					for _, eGrandparent := range eParent.Previous {
						uParent.Previous = append(uParent.Previous, eGrandparent)
					}

					chainToWrite.Parents = append(chainToWrite.Parents, uParent)
					eC.Parents = removeParent(eC.Parents, eParentIndex)
				}
			}
			if !parentMatch {
				chainToWrite.Parents = append(chainToWrite.Parents, nParent)
			}
		}

		for _, eParent := range eC.Parents {
			chainToWrite.Parents = append(chainToWrite.Parents, eParent)
		}

		chainToJson(chainToWrite, w.Name)
		w.Intake = 0
		w.Chain = chain{}
	}
}

// WorkersStats returns a slice of type WorkerStats
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

// nextMatch := false
// for i := 0; i < len(cParent.Next); i++ {
// 	cChild := &cParent.Next[i]

// 	cW := cChild.Word
// 	cV := cChild.Value
// 	for i := 0; i < len(eParent.Next); i++ {
// 		eChild := &eParent.Next

// 		eW := eChild.Word
// 		eV := eChild.Value

// 		if eW == cW {
// 			nextMatch = true
// 			eV += cV
// 			continue
// 		}
// 	}
// }

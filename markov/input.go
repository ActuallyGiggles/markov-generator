package markov

import (
	"strings"
)

// In adds an entry into a specific chain.
func In(chainName string, content string) {
	if content == "" || len(content) <= 0 {
		return
	}

	workerMapMx.Lock()
	w, ok := workerMap[chainName]
	workerMapMx.Unlock()

	if !ok {
		w = newWorker(chainName)
	}

	w.ChainMx.Lock()
	//debugLog("Input Locked")
	w.addInput(content)
	w.ChainMx.Unlock()
	//debugLog("Input Unlocked")

	writeCounter()
}

func (w *worker) addInput(content string) {
	slice := prepareContentForChainProcessing(content)
	extractStartAndSaveToChain(&w.Chain, w.Name, slice)
	extractAndSaveToChain(&w.Chain, w.Name, slice)

	w.Intake++
	stats.LifetimeInputs++
	stats.SessionInputs++
}

func prepareContentForChainProcessing(content string) []string {
	var returnSlice []string
	returnSlice = append(returnSlice, startKey)
	slice := strings.Split(content, " ")
	if len(slice) > 1 {
		for i := 0; i < len(slice)-1; i++ {
			firstWord := slice[i]
			secondWord := slice[i+1]
			returnSlice = append(returnSlice, firstWord+" "+secondWord)
		}
	} else {
		returnSlice = append(returnSlice, slice[0])
	}
	returnSlice = append(returnSlice, endKey)
	return returnSlice
}

func extractStartAndSaveToChain(c *chain, name string, slice []string) {
	start := slice[0]
	next := slice[1]

	parentExists := false
	for i := 0; i < len(c.Parents); i++ {
		parent := &c.Parents[i]
		if parent.Word == start {
			parentExists = true

			childExists := false
			for i := 0; i < len(parent.Children); i++ {
				child := &parent.Children[i]
				if child.Word == next {
					childExists = true
					child.Value += 1
				}
			}

			if !childExists {
				child := child{
					Word:  next,
					Value: 1,
				}
				parent.Children = append(parent.Children, child)
			}
		}
	}

	if !parentExists {
		var children []child
		child := child{
			Word:  next,
			Value: 1,
		}
		children = append(children, child)
		parent := parent{
			Word:     start,
			Children: children,
		}
		c.Parents = append(c.Parents, parent)
	}
}

func extractAndSaveToChain(c *chain, name string, slice []string) {
	for i := 0; i < len(slice)-2; i++ {
		current := slice[i+1]
		next := slice[i+2]

		parentExists := false
		for i := 0; i < len(c.Parents); i++ {
			parent := &c.Parents[i]
			if parent.Word == current {
				parentExists = true

				childExists := false
				for i := 0; i < len(parent.Children); i++ {
					child := &parent.Children[i]
					if child.Word == next {
						childExists = true

						child.Value += 1
					}
				}

				if !childExists {
					child := child{
						Word:  next,
						Value: 1,
					}
					parent.Children = append(parent.Children, child)
				}
			}
		}

		if !parentExists {
			var children []child
			child := child{
				Word:  next,
				Value: 1,
			}
			children = append(children, child)

			parent := parent{
				Word:     current,
				Children: children,
			}
			c.Parents = append(c.Parents, parent)
		}
	}
}

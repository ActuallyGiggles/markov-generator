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
	extractHead(&w.Chain, w.Name, slice)
	extractBody(&w.Chain, w.Name, slice)
	extractTail(&w.Chain, w.Name, slice)

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

func extractHead(c *chain, name string, slice []string) {
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

func extractBody(c *chain, name string, slice []string) {
	for i := 0; i < len(slice)-2; i++ {
		current := slice[i+1]
		next := slice[i+2]
		previous := slice[i]

		parentExists := false
		for i := 0; i < len(c.Parents); i++ {
			parent := &c.Parents[i]
			if parent.Word == current {
				parentExists = true

				// Deal with child
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

				// Deal with grandparent
				grandparentExists := false
				for i := 0; i < len(parent.Grandparents); i++ {
					grandparent := &parent.Grandparents[i]
					if grandparent.Word == previous {
						grandparentExists = true

						grandparent.Value += 1
					}
				}

				if !grandparentExists {
					grandparent := grandparent{
						Word:  previous,
						Value: 1,
					}
					parent.Grandparents = append(parent.Grandparents, grandparent)
				}
			}
		}

		if !parentExists {
			// Deal with child
			var children []child
			child := child{
				Word:  next,
				Value: 1,
			}
			children = append(children, child)

			// Deal with grandparent
			var grandparents []grandparent
			grandparent := grandparent{
				Word:  previous,
				Value: 1,
			}
			grandparents = append(grandparents, grandparent)

			// Add all to parent
			parent := parent{
				Word:         current,
				Children:     children,
				Grandparents: grandparents,
			}
			c.Parents = append(c.Parents, parent)
		}
	}
}

func extractTail(c *chain, name string, slice []string) {
	end := slice[len(slice)-1]
	previous := slice[len(slice)-2]

	parentExists := false
	for i := 0; i < len(c.Parents); i++ {
		parent := &c.Parents[i]
		if parent.Word == end {
			parentExists = true

			grandparentExists := false
			for i := 0; i < len(parent.Grandparents); i++ {
				grandparent := &parent.Grandparents[i]
				if grandparent.Word == previous {
					grandparentExists = true
					grandparent.Value += 1
				}
			}

			if !grandparentExists {
				grandparent := grandparent{
					Word:  previous,
					Value: 1,
				}
				parent.Grandparents = append(parent.Grandparents, grandparent)
			}
		}
	}

	if !parentExists {
		var grandparents []grandparent
		grandparent := grandparent{
			Word:  previous,
			Value: 1,
		}
		grandparents = append(grandparents, grandparent)
		parent := parent{
			Word:         end,
			Grandparents: grandparents,
		}
		c.Parents = append(c.Parents, parent)
	}
}

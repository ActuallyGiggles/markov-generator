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

	w.addInput(content)

	writeCounter()
}

func contentToChain(c *chain, name string, content string) {
	slice := prepareContentForChainProcessing(content)

	extractStartAndSaveToChain(c, name, slice)
	extractEndAndSaveToChain(c, name, slice)
	extractAndSaveToChain(c, name, slice)
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
			for i := 0; i < len(parent.Next); i++ {
				child := &parent.Next[i]
				if child.Word == next {
					childExists = true
					child.Value += 1
				}
			}

			if !childExists {
				child := word{
					Word:  next,
					Value: 1,
				}
				parent.Next = append(parent.Next, child)
			}
		}
	}

	if !parentExists {
		var children []word
		child := word{
			Word:  next,
			Value: 1,
		}
		children = append(children, child)
		parent := parent{
			Word: start,
			Next: children,
		}
		c.Parents = append(c.Parents, parent)
	}
}

func extractEndAndSaveToChain(c *chain, name string, slice []string) {
	end := slice[len(slice)-1]
	previous := slice[len(slice)-2]

	parentExists := false
	for i := 0; i < len(c.Parents); i++ {
		parent := &c.Parents[i]
		if parent.Word == end {
			parentExists = true

			grandparentExists := false
			for i := 0; i < len(parent.Previous); i++ {
				grandparent := &parent.Previous[i]

				if grandparent.Word == previous {
					grandparentExists = true
					grandparent.Value += 1
				}
			}

			if !grandparentExists {
				grandparent := word{
					Word:  previous,
					Value: 1,
				}
				parent.Previous = append(parent.Previous, grandparent)
			}
		}
	}

	if !parentExists {
		var grandparents []word
		grandparent := word{
			Word:  previous,
			Value: 1,
		}
		grandparents = append(grandparents, grandparent)
		parent := parent{
			Word:     end,
			Previous: grandparents,
		}
		c.Parents = append(c.Parents, parent)
	}
}

func extractAndSaveToChain(c *chain, name string, slice []string) {
	for i := 0; i < len(slice)-2; i++ {
		previous := slice[i]
		current := slice[i+1]
		next := slice[i+2]

		parentExists := false
		for i := 0; i < len(c.Parents); i++ {
			parent := &c.Parents[i]
			if parent.Word == current {
				parentExists = true

				childExists := false
				for i := 0; i < len(parent.Next); i++ {
					child := &parent.Next[i]
					if child.Word == next {
						childExists = true

						child.Value += 1
					}
				}

				if !childExists {
					child := word{
						Word:  next,
						Value: 1,
					}
					parent.Next = append(parent.Next, child)
				}

				grandparentExists := false
				for i := 0; i < len(parent.Previous); i++ {
					grandparent := &parent.Previous[i]
					if grandparent.Word == previous {
						grandparentExists = true

						grandparent.Value += 1
					}
				}

				if !grandparentExists {
					grandparent := word{
						Word:  previous,
						Value: 1,
					}
					parent.Previous = append(parent.Previous, grandparent)
				}
			}
		}

		if !parentExists {
			var children []word
			child := word{
				Word:  next,
				Value: 1,
			}
			children = append(children, child)

			var grandparents []word
			grandparent := word{
				Word:  previous,
				Value: 1,
			}
			grandparents = append(grandparents, grandparent)

			parent := parent{
				Word:     current,
				Next:     children,
				Previous: grandparents,
			}
			c.Parents = append(c.Parents, parent)
		}
	}
}

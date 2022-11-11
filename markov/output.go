package markov

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Out takes output instructions and returns an output and error.
func Out(oi OutputInstructions) (output string, err error) {
	name := oi.Chain
	method := oi.Method
	target := oi.Target

	defer duration(track("OUTPUT: " + name))

	if workerMap[name] == nil {
		return
	}

	workerMap[name].ChainMx.RLock()
	defer workerMap[name].ChainMx.RUnlock()

	switch method {
	case "LikelyBeginning":
		output, err = likelyBeginning(name)
	case "TargetedBeginning":
		output, err = targetedBeginning(name, target)
	}

	if err == nil {
		stats.LifetimeOutputs++
		stats.SessionOutputs++
	}

	return output, err
}

func likelyBeginning(name string) (output string, err error) {
	output = ""
	p := startKey
	c := ""

	for true {
		f, err := os.Open("./markov-chains/" + name + ".json")
		if err != nil {
			return "", err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		parentExists := false
		for dec.More() {
			var cP parent

			err = dec.Decode(&cP)
			if err != nil {
				fmt.Println(name)
				fmt.Println(cP)
				panic(err)
			}

			if cP.Word == p {
				parentExists = true

				c = getNextWord(cP)

				if c == endKey {
					parentSplit := strings.Split(p, " ")

					if len(parentSplit) == 1 {
						output = output + p
						return output, nil
					}

					output = output + parentSplit[1]
					return output, nil
				} else {
					cSplit := strings.Split(c, " ")
					output = output + cSplit[0] + " "

					p = c
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", p, name))
		}
	}

	// c, err := jsonToChain(name)
	// if err != nil {
	// 	return "", err
	// }

	// for true {
	// 	parentExists := false
	// 	for _, cParent := range c.Parents {
	// 		if cParent.Word == p {
	// 			parentExists = true
	// 			child = getNextWord(cParent)

	// 			if child == endKey {
	// 				parentSplit := strings.Split(p, " ")

	// 				if len(parentSplit) == 1 {
	// 					output = output + p
	// 					return output, nil
	// 				}

	// 				output = output + parentSplit[1]
	// 				return output, nil
	// 			} else {
	// 				childSplit := strings.Split(child, " ")
	// 				output = output + childSplit[0] + " "

	// 				p = child
	// 				continue
	// 			}
	// 		}
	// 	}

	// 	if !parentExists {
	// 		return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parent, name))
	// 	}
	// }

	return output, nil
}

func targetedBeginning(name string, target string) (output string, err error) {
	output = ""
	p := ""
	c := ""

	// c, err := jsonToChain(name)
	// if err != nil {
	// 	return "", err
	// }

	var initialList []string

	f, err := os.Open("./markov-chains/" + name + ".json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var cP parent

		err = dec.Decode(&cP)
		if err != nil {
			panic(err)
		}

		potentialParentSplit := strings.Split(cP.Word, " ")

		if potentialParentSplit[0] == target {
			initialList = append(initialList, cP.Word)
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	p = pickRandomParent(initialList)
	parentSplit := strings.Split(p, " ")
	output = parentSplit[0] + " "

	for true {
		f, err := os.Open("./markov-chains/" + name + ".json")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		parentExists := false
		for dec.More() {
			var cP parent

			err = dec.Decode(&cP)
			if err != nil {
				panic(err)
			}

			if cP.Word == p {
				parentExists = true

				c = getNextWord(cP)

				if c == endKey {
					parentSplit := strings.Split(p, " ")

					if len(parentSplit) == 1 {
						output = output + p
						return output, nil
					}

					output = output + parentSplit[1]
					return output, nil
				} else {
					cSplit := strings.Split(c, " ")
					output = output + cSplit[0] + " "

					p = c
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", p, name))
		}
	}

	// for true {
	// 	parentExists := false
	// 	for parentNumber, cParent := range c.Parents {
	// 		if initial {
	// 			if parentNumber >= len(c.Parents)-1 {
	// 				initial = false
	// 				parentExists = true
	// 				if len(initialList) <= 0 {
	// 					return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	// 				}
	// 				parent = pickRandomParent(initialList)
	// 				parentSplit := strings.Split(parent, " ")
	// 				output = parentSplit[0] + " "
	// 				break
	// 			}

	// 			potentialParentSplit := strings.Split(cParent.Word, " ")
	// 			if potentialParentSplit[0] == target {
	// 				initialList = append(initialList, cParent.Word)
	// 				continue
	// 			} else {
	// 				continue
	// 			}
	// 		}

	// 		if cParent.Word == parent {
	// 			parentExists = true
	// 			child = getNextWord(cParent)

	// 			if child == endKey {
	// 				parentSplit := strings.Split(parent, " ")

	// 				if len(parentSplit) == 1 {
	// 					output = output + parent
	// 					return output, nil
	// 				}

	// 				output = output + parentSplit[1]
	// 				return output, nil
	// 			} else {
	// 				childSplit := strings.Split(child, " ")
	// 				output = output + childSplit[0] + " "

	// 				parent = child
	// 				continue
	// 			}
	// 		}
	// 	}

	// 	if !parentExists {
	// 		return output, errors.New(fmt.Sprintf("%s does not contain parent: %s", name, parent))
	// 	}
	// }

	return output, nil
}

func getNextWord(parent parent) (child string) {
	var wrS []wRand
	for _, word := range parent.Children {
		w := word.Word
		v := word.Value
		item := wRand{
			Word:  w,
			Value: v,
		}
		wrS = append(wrS, item)
	}
	child = weightedRandom(wrS)

	return child
}

func pickRandomParent(parents []string) (parent string) {
	parent = pickRandomFromSlice(parents)

	return parent
}

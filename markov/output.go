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

	defer duration(track("output", name))

	// if zipping {
	// 	return "", errors.New("Currently zipping, try again later.")
	// }

	if exists := chainExists(name); !exists {
		fmt.Println(oi)
		return "", errors.New("Chain '" + name + "' is not found in directory.")
	}

	if exists := workerExists(name); exists {
		workerMap[name].ChainMx.RLock()
		defer workerMap[name].ChainMx.RUnlock()
	}

	switch method {
	case "LikelyBeginning":
		output, err = likelyBeginning(name)
	case "LikelyEnding":
		output, err = likelyEnding(name)
	case "TargetedBeginning":
		output, err = targetedBeginning(name, target)
	case "TargetedEnding":
		output, err = targetedEnding(name, target)
	case "TargetedMiddle":
		output, err = targetedMiddle(name, target)
	}

	if err == nil {
		stats.LifetimeOutputs++
		stats.SessionOutputs++
	}

	return output, err
}

func likelyBeginning(name string) (output string, err error) {
	var child string

	parentWord, err := getStartWord(name)
	if err != nil {
		return "", err
	}

	output = strings.Split(parentWord, " ")[0] + " "

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
			var currentParent parent

			err = dec.Decode(&currentParent)
			if err != nil {
				fmt.Println(name)
				fmt.Println(currentParent)
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				child = getNextWord(currentParent)

				if child == endKey {
					parentSplit := strings.Split(parentWord, " ")

					if len(parentSplit) == 1 {
						return output, nil
					}

					output = output + parentSplit[1]
					return output, nil
				} else {
					cSplit := strings.Split(child, " ")
					output = output + cSplit[0] + " "

					parentWord = child
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func likelyEnding(name string) (output string, err error) {
	var grandparent string

	parentWord, err := getEndWord(name)
	if err != nil {
		return "", err
	}

	if s := strings.Split(parentWord, " "); len(s) > 1 {
		output = s[1] + " "
	} else {
		output = s[0] + " "
	}

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
			var currentParent parent

			err = dec.Decode(&currentParent)
			if err != nil {
				fmt.Println(name)
				fmt.Println(currentParent)
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				grandparent = getPreviousWord(currentParent)

				if grandparent == startKey {
					parentSplit := strings.Split(parentWord, " ")

					if len(parentSplit) == 1 {
						return output, nil
					}

					output = parentSplit[0] + " " + output
					return output, nil
				} else {
					gSplit := strings.Split(grandparent, " ")
					output = gSplit[1] + " " + output

					parentWord = grandparent
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedBeginning(name, target string) (output string, err error) {
	var parentWord string
	var childChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_head.json")
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
		var currentParent child

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		potentialParentSplit := strings.Split(currentParent.Word, " ")

		if potentialParentSplit[0] == target {
			initialList = append(initialList, Choice{
				Word:   currentParent.Word,
				Weight: currentParent.Value,
			})
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	parentSplit := strings.Split(parentWord, " ")
	output = parentSplit[0] + " "

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
			var currentParent parent

			err = dec.Decode(&currentParent)
			if err != nil {
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				childChosen = getNextWord(currentParent)

				if childChosen == endKey {
					parentSplit := strings.Split(parentWord, " ")

					if len(parentSplit) == 1 {
						return output, nil
					}

					output = output + parentSplit[1]
					return output, nil
				} else {
					cSplit := strings.Split(childChosen, " ")
					output = output + cSplit[0] + " "

					parentWord = childChosen
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedEnding(name, target string) (output string, err error) {
	var parentWord string
	var grandparentChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_tail.json")
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
		var currentParent grandparent

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		potentialParentSplit := strings.Split(currentParent.Word, " ")

		if len(potentialParentSplit) == 1 {
			if potentialParentSplit[0] == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		} else {
			if potentialParentSplit[1] == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	parentSplit := strings.Split(parentWord, " ")
	if len(parentSplit) == 1 {
		output = parentSplit[0] + " "
	} else {
		output = parentSplit[1] + " "
	}

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
			var currentParent parent

			err = dec.Decode(&currentParent)
			if err != nil {
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				grandparentChosen = getPreviousWord(currentParent)

				if grandparentChosen == startKey {
					parentSplit := strings.Split(parentWord, " ")

					if len(parentSplit) == 1 {
						return output, nil
					}

					output = parentSplit[0] + " " + output
					return output, nil
				} else {
					gSplit := strings.Split(grandparentChosen, " ")
					output = gSplit[1] + " " + output

					parentWord = grandparentChosen
					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedMiddle(name, target string) (output string, err error) {
	var parentWord string
	var childChosen string
	var grandparentChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_body.json")
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
		var currentParent parent

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		potentialParentSplit := strings.Split(currentParent.Word, " ")

		if len(potentialParentSplit) == 2 {
			if potentialParentSplit[0] == target || potentialParentSplit[1] == target {
				goto addParent
			} else {
				continue
			}
		} else {
			if potentialParentSplit[0] == target {
				goto addParent
			} else {
				continue
			}
		}

	addParent:
		var totalWeight int

		for _, child := range currentParent.Children {
			totalWeight += child.Value
		}

		for _, grandparent := range currentParent.Grandparents {
			totalWeight += grandparent.Value
		}

		initialList = append(initialList, Choice{
			Word:   currentParent.Word,
			Weight: totalWeight,
		})
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}

	firstParentWord := parentWord

	var forwardComplete bool

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
			var currentParent parent

			err = dec.Decode(&currentParent)
			if err != nil {
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				// Build forwards
				if !forwardComplete {
					goto forward
				} else {

					goto backward
				}

			forward:
				childChosen = getNextWord(currentParent)
				if childChosen == endKey {
					if parentSplit := strings.Split(parentWord, " "); len(parentSplit) == 1 {
						output += parentSplit[0]
					} else {
						output += parentSplit[1]
					}

					forwardComplete = true

					// Forwards is done, move onto building backwards
					parentWord = firstParentWord
					continue
				} else {
					cSplit := strings.Split(childChosen, " ")
					output += cSplit[0] + " "
					parentWord = childChosen

					continue
				}

			backward:
				// Build backwards
				grandparentChosen = getPreviousWord(currentParent)
				if grandparentChosen == startKey {
					parentSplit := strings.Split(parentWord, " ")

					if len(parentSplit) == 1 {
						goto end
					}

					output = parentSplit[0] + " " + output

					goto end
				} else {
					gSplit := strings.Split(grandparentChosen, " ")
					output = gSplit[1] + " " + output
					parentWord = grandparentChosen

					continue
				}
			}
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

end:
	return output, nil
}

func getNextWord(parent parent) (child string) {
	var wrS []Choice
	for _, word := range parent.Children {
		w := word.Word
		v := word.Value
		item := Choice{
			Word:   w,
			Weight: v,
		}
		wrS = append(wrS, item)
	}
	child, _ = weightedRandom(wrS)

	return child
}

func getPreviousWord(parent parent) (grandparent string) {
	var wrS []Choice
	for _, word := range parent.Grandparents {
		w := word.Word
		v := word.Value
		item := Choice{
			Word:   w,
			Weight: v,
		}
		wrS = append(wrS, item)
	}
	grandparent, _ = weightedRandom(wrS)

	return grandparent
}

func pickRandomParent(parents []string) (parent string) {
	parent = pickRandomFromSlice(parents)

	return parent
}

func getStartWord(name string) (phrase string, err error) {
	var sum int

	f, err := os.Open("./markov-chains/" + name + "_head.json")
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var child child

		err = dec.Decode(&child)
		if err != nil {
			fmt.Println(name)
			fmt.Println(child)
			panic(err)
		}

		sum += child.Value
	}

	f.Close()

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err = os.Open("./markov-chains/" + name + "_head.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec = json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var child child

		err = dec.Decode(&child)
		if err != nil {
			fmt.Println(name)
			fmt.Println(child)
			panic(err)
		}

		r -= child.Value

		if r < 0 {
			return child.Word, nil
		}
	}

	return "", errors.New("Internal error - code should not reach this point")
}

func getEndWord(name string) (phrase string, err error) {
	var sum int

	f, err := os.Open("./markov-chains/" + name + "_tail.json")
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var grandparent grandparent

		err = dec.Decode(&grandparent)
		if err != nil {
			fmt.Println(name)
			fmt.Println(grandparent)
			panic(err)
		}

		sum += grandparent.Value
	}

	f.Close()

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err = os.Open("./markov-chains/" + name + "_tail.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec = json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var grandparent grandparent

		err = dec.Decode(&grandparent)
		if err != nil {
			fmt.Println(name)
			fmt.Println(grandparent)
			panic(err)
		}

		r -= grandparent.Value

		if r < 0 {
			return grandparent.Word, nil
		}
	}

	return "", errors.New("Internal error - code should not reach this point")
}

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

	if exists := chainExists(name); !exists {
		return "", errors.New("Chain '" + name + "' is not found in directory.")
	}

	if exists := workerExists(name); exists {
		workerMap[name].ChainMx.RLock()
		defer workerMap[name].ChainMx.RUnlock()
	}

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
	c := ""

	p, err := getStartWord(name)
	if err != nil {
		return "", err
	}

	output = strings.Split(p, " ")[0] + " "

	for true {
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

	return output, nil
}

func targetedBeginning(name string, target string) (output string, err error) {
	output = ""
	p := ""
	c := ""

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
		var cP child

		err = dec.Decode(&cP)
		if err != nil {
			panic(err)
		}

		potentialParentSplit := strings.Split(cP.Word, " ")

		if potentialParentSplit[0] == target {
			initialList = append(initialList, Choice{
				Word:   cP.Word,
				Weight: cP.Value,
			})
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	p, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	parentSplit := strings.Split(p, " ")
	output = parentSplit[0] + " "

	for true {
		f, err := os.Open("./markov-chains/" + name + "_tail.json")
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

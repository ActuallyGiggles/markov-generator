package markov

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jmcvetta/randutil"
)

func debugLog(v ...any) {
	if debug {
		log.Println(v...)
	}
}

func chains() []string {
	files, err := ioutil.ReadDir("./markov-chains/")
	var s []string
	if err != nil {
		return s
	}
	for _, file := range files {
		s = append(s, strings.TrimSuffix(file.Name(), ".json"))
	}
	return s
}

func now() string {
	return time.Now().Format("15:04:05")
}

func jsonToChain(name string) (c chain, err error) {
	path := "./markov-chains/" + name + ".json"
	file, err := os.Open(path)
	if err != nil {
		debugLog("Failed reading file:", err)
		return chain{}, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		debugLog("Error when unmarshalling file:", path, "\n", err)
		return chain{}, err
	}

	return c, nil
}

func chainToJson(c chain, name string) {
	path := "./markov-chains/" + name + ".json"

	chainData, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		debugLog(err)
	}

	f, err := os.Create(path)
	if err != nil {
		debugLog(err)
	}

	fileStat, ferr := f.Stat()

	n2, err := f.Write(chainData)
	f.Close()
	if err != nil {
		debugLog(err)
	}

	if ferr == nil {
		fileSize := fileStat.Size()
		change := int64(n2) - fileSize

		if change > 0 {
			debugLog("wrote successfully to", path)
			debugLog(int64(n2), "-", fileSize, "=", change)
		} else {
			fmt.Println("wrote unsuccessfully to", path)
			fmt.Println(int64(n2), "-", fileSize, "=", change)
			debugLog("wrote unsuccessfully to", path)
			debugLog(int64(n2), "-", fileSize, "=", change)
		}
	} else {
		debugLog("wrote successfully to", path)
		debugLog(int64(n2))
	}
}

func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

func track(chain string) (string, time.Time) {
	return chain, time.Now()
}

func duration(chain string, start time.Time) {
	debugLog(chain + ": " + fmt.Sprint(time.Since(start)))
}

// CurrentChains returns the names of all chains that have been made
func CurrentChains() []string {
	workerMapMx.Lock()
	var s []string
	for chain := range workerMap {
		s = append(s, chain)
	}
	workerMapMx.Unlock()
	return s
}

// WriteMode returns what the current mode is
func WriteMode() (mode string) {
	return writeMode
}

// TimeUntilWrite returns the duration until the next write cycle
func TimeUntilWrite() time.Duration {
	return nextWriteTime.Sub(time.Now())
}

// NextWriteTime returns what time the next write cycle will happen
func NextWriteTime() time.Time {
	return nextWriteTime
}

// PeakChainIntake returns the highest intake across all workers per session and at what time it happened
func PeakChainIntake() struct {
	Chain  string
	Amount int
	Time   time.Time
} {
	return peakChainIntake
}

func weightedRandom(itemsAndWeights []wRand) string {
	var choices []randutil.Choice

	for _, item := range itemsAndWeights {
		word := item.Word
		value := item.Value
		choices = append(choices, randutil.Choice{Weight: value, Item: word})
	}

	choice, err := randutil.WeightedChoice(choices)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%v", choice.Item)

	// // Create variable for slice of choice struct
	// var choices []wr.Choice

	// for _, item := range itemsAndWeights { // For every child, value in map
	// 	word := item.Word
	// 	value := item.Value
	// 	choices = append(choices, wr.Choice{Item: word, Weight: uint(value)}) // Add item, value to choices
	// }

	// chooser, err := wr.NewChooser(choices...) // Initialize chooser
	// if err != nil {
	// 	panic(err)
	// }

	// return chooser.Pick().(string) // Choose
}

func createChainsFolder() {
	// Create or check if markov markov db folder exists
	_, dberr := os.Stat("./markov-chains")
	if os.IsNotExist(dberr) {
		err := os.MkdirAll("./markov-chains", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func RandomNumber(min int, max int) (num int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	num = r.Intn(max-min) + min
	return num
}

func PickRandomFromSlice(slice []string) string {
	return slice[RandomNumber(0, len(slice))]
}

func removeCorGP(s []word, i int) []word {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeParent(s []parent, i int) []parent {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

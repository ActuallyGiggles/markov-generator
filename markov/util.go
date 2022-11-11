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

// CurrentChains returns the names of all chains that have been made.
func CurrentChains() []string {
	workerMapMx.Lock()
	var s []string
	for chain := range workerMap {
		s = append(s, chain)
	}
	workerMapMx.Unlock()
	return s
}

// WriteMode returns what the current mode is.
func WriteMode() (mode string) {
	return writeMode
}

// TimeUntilWrite returns the duration until the next write cycle.
func TimeUntilWrite() time.Duration {
	return stats.NextWriteTime.Sub(time.Now())
}

// NextWriteTime returns what time the next write cycle will happen.
func NextWriteTime() time.Time {
	return stats.NextWriteTime
}

// PeakIntake returns the highest intake across all workers per session and at what time it happened.
func PeakIntake() PeakIntakeStruct {
	return stats.PeakChainIntake
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
}

func createFolders() {
	_, err := os.Stat("./markov-chains")
	if os.IsNotExist(err) {
		err := os.MkdirAll("./markov-chains", 0755)
		if err != nil {
			panic(err)
		}
	}

	_, err = os.Stat("./markov-chains/stats")
	if os.IsNotExist(err) {
		err := os.MkdirAll("./markov-chains/stats", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func randomNumber(min int, max int) (num int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	num = r.Intn(max-min) + min
	return num
}

func pickRandomFromSlice(slice []string) string {
	return slice[randomNumber(0, len(slice))]
}

func removeChild(s []child, i int) []child {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeParent(s []parent, i int) []parent {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

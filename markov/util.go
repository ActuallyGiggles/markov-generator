package markov

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-json"

	wr "github.com/mroth/weightedrand"
)

func debugLog(v ...any) {
	if Debug {
		log.Println(v...)
	}
}

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func jsonToChain(path string) (chain map[string]map[string]map[string]int, exists bool) {
	file, err := os.Open(path)
	if err != nil {
		debugLog("Failed reading file:", err)
		return nil, false
	}
	defer file.Close()

	// /*content, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	log.Println("jsonToChain error: ", path, "\n", err)
	// 	return nil, false
	// }

	// err = json.Unmarshal(content, &chain)
	// if err != nil {
	// 	log.Println("Error when unmarshalling file:", path, "\n", err)
	// 	return nil, false
	// }*/

	err = json.NewDecoder(file).Decode(&chain)
	if err != nil {
		debugLog("Error when unmarshalling file:", path, "\n", err)
		return nil, false
	}

	return chain, true
}

func chainToJson(chain map[string]map[string]map[string]int, path string) {
	byteArray, err := GetBytes(chain)
	if err != nil {
		log.Panic(err)
	}
	err = os.WriteFile("file.txt", byteArray, 0644)
	if err != nil {
		log.Panic(err)
	}

	// file, _ := json.MarshalIndent(chain, "", " ")
	// _ = ioutil.WriteFile(path, file, 0644)

	/*chainData, err := json.MarshalIndent(chain, "", " ")
	if err != nil {
		debugLog("ERROR MARSHALLING ", path)
	}

	f, err := os.Create(path)
	if err != nil {
		debugLog("ERROR CREATING ", path, err)
	}

	n2, err := f.Write(chainData)
	f.Close()
	if err != nil {
		debugLog("ERROR WRITING ", path, err)
	}

	debugLog("wrote", n2, "bytes to", path)*/
}

// PrettyPrint prints out an object in a pretty format
func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

// createChains creates a markovdb folder if it doesn't exist.
func createChainsFolder() {
	// Create or check if main markov db folder exists
	_, dberr := os.Stat("markov/chains")
	if os.IsNotExist(dberr) {
		err := os.MkdirAll("markov/chains", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func now() string {
	return time.Now().Format("15:04:05")
}

// TimeUntilWrite returns the duration until the next write cycle
func TimeUntilWrite() time.Duration {
	return nextWriteTime.Sub(time.Now())
}

// NextWriteTime returns what time the next write cycle will happen
func NextWriteTime() time.Time {
	return nextWriteTime
}

// ChainPeakIntake returns the highest intake across all workers per session and at what time it happened
func ChainPeakIntake() struct {
	Chain  string
	Amount int
	Time   time.Time
} {
	return chainPeakIntake
}

func weightedRandom(itemsAndWeights map[string]int) string {
	// Create variable for slice of choice struct
	var choices []wr.Choice

	for item, value := range itemsAndWeights { // For every child, value in map
		choices = append(choices, wr.Choice{Item: item, Weight: uint(value)}) // Add item, value to choices
	}

	chooser, _ := wr.NewChooser(choices...) // Initialize chooser
	return chooser.Pick().(string)          // Choose
}

func doesSliceContainIndex(slice []string, index int) bool {
	if len(slice) > index {
		return true
	} else {
		return false
	}
}

// Chains returns the names of all chains that have been made
func chains() []string {
	files, err := ioutil.ReadDir("./markov/chains/")
	var s []string
	if err != nil {
		// fmt.Println("pass")
		return s
	}
	for _, file := range files {
		s = append(s, strings.TrimSuffix(file.Name(), ".json"))
	}
	return s
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

func track(chain string) (string, time.Time) {
	return chain, time.Now()
}

func duration(chain string, start time.Time) {
	debugLog(chain + ": " + fmt.Sprint(time.Since(start)))
}

// WriteMode returns what the current mode is
func WriteMode() (mode string) {
	return writeMode
}

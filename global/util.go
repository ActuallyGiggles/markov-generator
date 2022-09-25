package global

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func UpdateResourceLists() {
	for i := 0; i < len(Resources); i++ {
		if Resources[i].DiscordChannelName == "banned-users" {
			BannedUsers = strings.Split(Resources[i].Content, " ")
			for j, username := range BannedUsers {
				if username == "" || username == " " {
					BannedUsers = FastRemove(BannedUsers, j)
				}
			}
		}
		if Resources[i].DiscordChannelName == "regex" {
			r := strings.TrimSpace(Resources[i].Content)
			r = strings.ReplaceAll(r, " ", "|")
			Regex = regexp.MustCompile(r)
		}
	}
}

// FastRemove removes an index from a slice of strings without maintaining order
func FastRemove(s []string, i int) []string {
	s[i] = s[len(s)-1] // Copy last element to index i
	s = s[:len(s)-1]   // Truncate slice
	return s
}

// PrettyPrint prints out a map in a pretty format.
func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
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

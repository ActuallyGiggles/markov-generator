package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	updatingEmotes bool
)

func getEmotes() {
	fmt.Println()
	fmt.Println("Gathering emotes...")
	updatingEmotes = true
	GetBroadcasters()
	TGlobal := getTwitchGlobalEmotes()
	SevenGlobal := get7tvGlobalEmotes()
	BGlobal := getBttvGlobalEmotes()
	FGlobal := getFfzGlobalEmotes()
	TChannel := getTwitchChannelEmotes()
	SevenChannel := get7tvChannelEmotes()
	BChannel := getBttvChannelEmotes()
	FChannel := getFfzChannelEmotes()
	updatingEmotes = false

	globals := TGlobal + SevenGlobal + BGlobal + FGlobal

	fmt.Println("[Global]:", globals)
	fmt.Println("\tTwitch Global:", TGlobal)
	fmt.Println("\t7tv Global:", SevenGlobal)
	fmt.Println("\tBTTV Global:", BGlobal)
	fmt.Println("\tFFZ Global:", FGlobal)
	fmt.Println("[Channel]")
	for _, channel := range Config.Channels {
		fmt.Printf("\t[%s]: %d\n", channel, TChannel[channel]+SevenChannel[channel]+BChannel[channel]+FChannel[channel])
		fmt.Println("\t\tTwitch:", TChannel[channel])
		fmt.Println("\t\t7tv:", SevenChannel[channel])
		fmt.Println("\t\tBTTV:", BChannel[channel])
		fmt.Println("\t\tFFZ:", FChannel[channel])
	}
	fmt.Println()

	go func() {
		for range time.Tick(1 * time.Hour) {
			updateEmotes()
		}
	}()
}

func updateEmotes() {
	updatingEmotes = true
	ChannelEmotes = nil
	getTwitchChannelEmotes()
	get7tvChannelEmotes()
	getBttvChannelEmotes()
	getFfzChannelEmotes()
	updatingEmotes = false
	fmt.Println("Emotes updated.")
}

func GetBroadcasters() {
	for i := 0; i < len(Users); i++ {
		user := &Users[i]
		url := "https://api.twitch.tv/helix/users?login=" + user.Name

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", "Bearer "+Config.OAuth)
		req.Header.Set("Client-Id", Config.ClientID)
		if err != nil {
			log.Println("GetBroadcasterID failed\n", err.Error())
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("GetBroadcasterID failed\n", err.Error())
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			log.Printf("GetBroadcaters(%s) is not OK\n\t%s", user.Name, string(body))
		}
		broadcaster := Broadcaster[Data]{}
		if err := json.Unmarshal(body, &broadcaster); err != nil {
			log.Println("GetBroadcasterID failed\n", err.Error())
		}
		for _, v := range broadcaster.Data {
			user.ID = v.ID
		}
	}
}

func getTwitchGlobalEmotes() (number int) {
	url := "https://api.twitch.tv/helix/chat/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+Config.OAuth)
	req.Header.Set("Client-Id", Config.ClientID)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := TwitchEmoteAPIResponse[TwitchGlobalEmote]{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		panic(err)
	}

	for _, emote := range emotes.Data {
		GlobalEmotes = append(GlobalEmotes, emote.Name)
		number++
	}

	return number
}

func getTwitchChannelEmotes() map[string]int {
	c := make(map[string]int)
	for i := 0; i < len(Users); i++ {
		user := &Users[i]
		url := "https://api.twitch.tv/helix/chat/emotes?broadcaster_id=" + user.ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", "Bearer "+Config.OAuth)
		req.Header.Set("Client-Id", Config.ClientID)
		if err != nil {
			log.Printf("\t getTwitchChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", user.Name)
			log.Println(err.Error())
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("getTwitchChannelEmotes failed\n", err.Error())
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := TwitchEmoteAPIResponse[TwitchChannelEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			log.Println("getTwitchChannelEmotes failed\n", err.Error())
		}

		var number int

		for _, emote := range emotes.Data {
			ChannelEmotes = append(ChannelEmotes, emote.Name)
			number++
		}

		c[user.Name] = number
	}

	return c
}

func get7tvGlobalEmotes() (number int) {
	url := "https://api.7tv.app/v2/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []SevenTVEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		log.Fatal(err)
	}

	for _, emote := range emotes {
		GlobalEmotes = append(GlobalEmotes, emote.Name)
		number++
	}

	return number
}

func get7tvChannelEmotes() map[string]int {
	c := make(map[string]int)
	for i := 0; i < len(Users); i++ {
		user := &Users[i]
		url := "https://api.7tv.app/v2/users/" + user.Name + "/emotes"

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			panic(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := []SevenTVEmote{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			panic(err)
		}

		var number int

		for _, emote := range emotes {
			ChannelEmotes = append(ChannelEmotes, emote.Name)
			number++
		}

		c[user.Name] = number
	}
	return c
}

func getBttvGlobalEmotes() (number int) {
	url := "https://api.betterttv.net/3/cached/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []BttvEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		panic(err)
	}

	for _, emote := range emotes {
		GlobalEmotes = append(GlobalEmotes, emote.Name)
		number++
	}
	return number
}

func getBttvChannelEmotes() map[string]int {
	c := make(map[string]int)
	for i := 0; i < len(Users); i++ {
		user := &Users[i]
		url := "https://api.betterttv.net/3/cached/users/twitch/" + user.ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			panic(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := BttvChannelEmotes[BttvEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			panic(err)
		}

		var number int

		for _, emote := range emotes.ChannelEmotes {
			ChannelEmotes = append(ChannelEmotes, emote.Name)
			number++
		}
		for _, emote := range emotes.SharedEmotes {
			ChannelEmotes = append(ChannelEmotes, emote.Name)
			number++
		}
		c[user.Name] = number
	}
	return c
}

func getFfzGlobalEmotes() (number int) {
	url := "https://api.frankerfacez.com/v1/set/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	set := FfzSets{}
	if err := json.Unmarshal(body, &set); err != nil {
		panic(err)
	}

	for _, emotes := range set.Sets {
		for _, emote := range emotes.Emoticons {
			GlobalEmotes = append(GlobalEmotes, emote.Name)
			number++
		}
	}

	return number
}

func getFfzChannelEmotes() map[string]int {
	c := make(map[string]int)
	for i := 0; i < len(Users); i++ {
		user := &Users[i]
		url := "https://api.frankerfacez.com/v1/room/id/" + user.ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			panic(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		set := FfzSets{}
		if err := json.Unmarshal(body, &set); err != nil {
			panic(err)
		}

		var number int

		for _, emotes := range set.Sets {
			for _, emote := range emotes.Emoticons {
				ChannelEmotes = append(ChannelEmotes, emote.Name)
				number++
			}
		}

		c[user.Name] = number
	}
	return c
}

// func GetLiveStatuses() {
// 	for i := 0; i < len(Users); i++ {
// 		user := &Users[i]
// 		url := "https://api.twitch.tv/helix/streams?user_login=" + user.Name
// 		var jsonStr = []byte(`{"":""}`)
// 		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
// 		req.Header.Set("Authorization", "Bearer "+Config.OAuth)
// 		req.Header.Set("Client-Id", Config.ClientID)
// 		if err != nil {
// 			log.Println(err.Error())
// 		}
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			log.Println(err.Error())
// 			return
// 		}
// 		defer resp.Body.Close()
// 		body, _ := ioutil.ReadAll(resp.Body)
// 		var stream StreamStatusData
// 		if err := json.Unmarshal(body, &stream); err != nil {
// 			log.Println(err.Error())
// 		}
// 		if len(stream.Data) == 0 {
// 			user.IsLive = false
// 		} else {
// 			user.IsLive = true
// 		}
// 	}

// 	for range time.Tick(1 * time.Minute) {
// 		for i := 0; i < len(Users); i++ {
// 			user := &Users[i]
// 			url := "https://api.twitch.tv/helix/streams?user_login=" + user.Name
// 			var jsonStr = []byte(`{"":""}`)
// 			req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
// 			req.Header.Set("Authorization", "Bearer "+Config.OAuth)
// 			req.Header.Set("Client-Id", Config.ClientID)
// 			if err != nil {
// 				log.Println(err.Error())
// 			}
// 			client := &http.Client{}
// 			resp, err := client.Do(req)
// 			if err != nil {
// 				log.Println(err.Error())
// 				return
// 			}
// 			defer resp.Body.Close()
// 			body, _ := ioutil.ReadAll(resp.Body)
// 			var stream StreamStatusData
// 			if err := json.Unmarshal(body, &stream); err != nil {
// 				log.Println(err.Error())
// 			}
// 			if len(stream.Data) == 0 {
// 				user.IsLive = false
// 			} else {
// 				user.IsLive = true
// 			}
// 		}
// 	}
// }

func GetLiveStatuses() {
	for range time.Tick(30 * time.Second) {
		for i := 0; i < len(Users); i++ {
			user := &Users[i]
			url := "https://api.twitch.tv/helix/streams?user_login=" + user.Name
			var jsonStr = []byte(`{"":""}`)
			req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Authorization", "Bearer "+Config.OAuth)
			req.Header.Set("Client-Id", Config.ClientID)
			if err != nil {
				log.Println(err.Error())
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err.Error())
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			var stream StreamStatusData
			if err := json.Unmarshal(body, &stream); err != nil {
				log.Println(err.Error())
			}
			if len(stream.Data) == 0 {
				user.IsLive = false
			} else {
				user.IsLive = true
			}
		}
	}
}

package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/terminal"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// GetTwitchBroadcasterID gets a broadcaster's twitch ID.
//
// Returns the ID and whether function was successful.
func GetBroadcasterInfo(channel string) (data Data, ok bool) {
	url := "https://api.twitch.tv/helix/users?login=" + channel

	d := Data{}
	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
	req.Header.Set("Client-Id", global.TwitchClientID)
	if err != nil {
		log.Println("	GetBroadcasterID failed\n", err)
		return d, false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("	GetBroadcasterID failed\n", err)
		return d, false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	broadcaster := Broadcaster[Data]{}
	if err := json.Unmarshal(body, &broadcaster); err != nil {
		log.Println("	GetBroadcasterID failed\n", err)
		return d, false
	}
	for _, v := range broadcaster.Data {
		d.ID = v.ID
		d.Login = v.Login
		d.DisplayName = v.DisplayName
		d.Type = v.Type
		d.BroadcasterType = v.BroadcasterType
		d.Description = v.Description
		d.ProfileImageUrl = v.ProfileImageUrl
		d.OfflineImageUrl = v.OfflineImageUrl
		d.ViewCount = v.ViewCount
		d.Email = v.Email
		d.CreatedAt = v.CreatedAt
	}
	return d, true
}

func getBroadcasterIDs() {
	for _, channel := range global.Directives {
		data, ok := GetBroadcasterInfo(channel.ChannelName)
		if ok {
			broadcaster[channel.ChannelName] = data
		}
	}
}

func getTwitchGlobalEmotes() {
	url := "https://api.twitch.tv/helix/chat/emotes/global"

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
	req.Header.Set("Client-Id", global.TwitchClientID)
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
		global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
	}
}

func getTwitchChannelEmotes() {
	for _, d := range broadcaster {
		ID := d.ID
		url := "https://api.twitch.tv/helix/chat/emotes?broadcaster_id=" + ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
		req.Header.Set("Client-Id", global.TwitchClientID)
		if err != nil {
			log.Printf("\t getTwitchChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", ID)
			log.Println(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("\t getTwitchChannelEmotes failed\n")
			log.Printf("\t For channel %s\n2", ID)
			log.Println(err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := TwitchEmoteAPIResponse[TwitchChannelEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			log.Printf("\t getTwitchChannelEmotes failed\n")
			log.Printf("\t For channel %s\n3", ID)
			log.Println(err)
		}

		for _, emote := range emotes.Data {
			global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
		}
	}
}

func get7tvGlobalEmotes() {
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
		global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
	}
}

func get7tvChannelEmotes() {
	for _, channel := range global.Directives {
		url := "https://api.7tv.app/v2/users/" + channel.ChannelName + "/emotes"

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Printf("\t get7tvChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", channel.ChannelName)
			log.Println(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("\t get7tvChannelEmotes failed\n")
			log.Printf("\t For channel %s\n2", channel.ChannelName)
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := []SevenTVEmote{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			continue
		}

		count := 0

		for _, emote := range emotes {
			thirdPartyChannelEmotesToUpdate[channel.ChannelName] = append(thirdPartyChannelEmotesToUpdate[channel.ChannelName], emote.Name)
			count += 1
			// global.ThirdPartyChannelEmotes[channel] = append(global.ThirdPartyChannelEmotes[channel], emote.Name)
			// global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
		}
	}
}

func getBttvGlobalEmotes() {
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
		global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
	}
}

func getBttvChannelEmotes() {
	for _, d := range broadcaster {
		ID := d.ID
		user := d.Login
		url := "https://api.betterttv.net/3/cached/users/twitch/" + ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Printf("\t getBttvChannelEmotes failed\n")
			log.Printf("\t For channel %s\n3", ID)
			log.Println(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("\t getBttvChannelEmotes failed\n")
			log.Printf("\t For channel %s\n3", ID)
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := BttvChannelEmotes[BttvEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			log.Printf("\t getBttvChannelEmotes failed\n")
			log.Printf("\t For channel %s\n3", ID)
			log.Println(err)
		}

		count := 0

		for _, emote := range emotes.ChannelEmotes {
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
			count += 1
		}
		for _, emote := range emotes.SharedEmotes {
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
			count += 1
		}
	}
}

func getFfzGlobalEmotes() {
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
	set := FfzEmotes{}
	if err := json.Unmarshal(body, &set); err != nil {
		panic(err)
	}

	for _, emotes := range set.Sets {
		for _, emote := range emotes.Emoticons {
			global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
		}
	}
}

func getFfzChannelEmotes() {
	for _, d := range broadcaster {
		ID := d.ID
		user := d.Login
		url := "https://api.frankerfacez.com/v1/room/id/" + ID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Printf("\t getFfzChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", ID)
			log.Println(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("\t getFfzChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", ID)
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		set := FfzEmotes{}
		if err := json.Unmarshal(body, &set); err != nil {
			log.Printf("\t getFfzChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", ID)
			log.Println(err)
		}

		count := 0

		for _, emotes := range set.Sets {
			for _, emote := range emotes.Emoticons {
				// global.ThirdPartyChannelEmotes[user] = append(global.ThirdPartyChannelEmotes[user], emote.Name)
				thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
				count += 1
			}
		}
	}
}

func GetLiveStatus(channelName string) (live bool) {
	url := "https://api.twitch.tv/helix/streams?user_login=" + channelName
	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
	req.Header.Set("Client-Id", global.TwitchClientID)
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
	var stream StreamStatusData
	if err := json.Unmarshal(body, &stream); err != nil {
		panic(err)
	}
	if len(stream.Data) == 0 {
		return false
	} else {
		return true
	}
}

func GetLiveStatuses() {
	IsLiveMx.Lock()
	defer IsLiveMx.Unlock()
	for _, channel := range global.Directives {
		r := GetLiveStatus(channel.ChannelName)
		IsLive[channel.ChannelName] = r
	}
	terminal.UpdateTerminal("live")
}

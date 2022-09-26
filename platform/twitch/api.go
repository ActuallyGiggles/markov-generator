package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/terminal"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetTwitchBroadcasterID gets a broadcaster's twitch ID.
//
// Returns the ID and whether function was successful.
func GetBroadcasterID(channel string) (ID string, ok bool) {
	url := "https://api.twitch.tv/helix/users?login=" + channel

	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
	req.Header.Set("Client-Id", global.TwitchClientID)
	if err != nil {
		fmt.Println("	GetBroadcasterID failed\n", err)
		return "", false
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("	GetBroadcasterID failed\n", err)
		return "", false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	broadcaster := TwitchGetBroadcasterID[Broadcaster]{}
	if err := json.Unmarshal(body, &broadcaster); err != nil {
		fmt.Println("	GetBroadcasterID failed\n", err)
		return "", false
	}
	for _, value := range broadcaster.Data {
		return value.ID, true
	}
	return "", false
}

func getBroadcasterIDs() {
	for _, channel := range global.Directives {
		id, ok := GetBroadcasterID(channel.ChannelName)
		if ok {
			broadcasterIDs[channel.ChannelName] = id
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
	for _, broadcasterID := range broadcasterIDs {
		url := "https://api.twitch.tv/helix/chat/emotes?broadcaster_id=" + broadcasterID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
		req.Header.Set("Client-Id", global.TwitchClientID)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n1", broadcasterID)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n2", broadcasterID)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := TwitchEmoteAPIResponse[TwitchChannelEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n3", broadcasterID)
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
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	emotes := []SevenTVEmote{}
	if err := json.Unmarshal(body, &emotes); err != nil {
		panic(err)
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
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n1", channel.ChannelName)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n2", channel.ChannelName)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := []SevenTVEmote{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			//fmt.Printf("\t %s not registered for 7tv.\n", channel.ChannelName)
			continue
		}

		count := 0

		for _, emote := range emotes {
			thirdPartyChannelEmotesToUpdate[channel.ChannelName] = append(thirdPartyChannelEmotesToUpdate[channel.ChannelName], emote.Name)
			count += 1
			// global.ThirdPartyChannelEmotes[channel] = append(global.ThirdPartyChannelEmotes[channel], emote.Name)
			// global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
		}

		//fmt.Printf("%s has %d 7TV emotes.", channel.PlatformChannelName, count)
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
	for user, broadcasterID := range broadcasterIDs {
		url := "https://api.betterttv.net/3/cached/users/twitch/" + broadcasterID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n1", broadcasterID)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n2", broadcasterID)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		emotes := BttvChannelEmotes[BttvEmote]{}
		if err := json.Unmarshal(body, &emotes); err != nil {
			fmt.Printf("\t %s not registered for Bttv.\n", user)
		}

		count := 0

		for _, emote := range emotes.ChannelEmotes {
			// global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
			// global.ThirdPartyChannelEmotes[user] = append(global.ThirdPartyChannelEmotes[user], emote.Name)
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
			count += 1
		}
		for _, emote := range emotes.SharedEmotes {
			// global.GlobalEmotes = append(global.GlobalEmotes, emote.Name)
			// global.ThirdPartyChannelEmotes[user] = append(global.ThirdPartyChannelEmotes[user], emote.Name)
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
			count += 1
		}

		//fmt.Printf("%s has %d BTTV emotes.", user, count)
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
	for user, broadcasterID := range broadcasterIDs {
		url := "https://api.frankerfacez.com/v1/room/id/" + broadcasterID

		var jsonStr = []byte(`{"":""}`)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n1", broadcasterID)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("\t For channel %s\n2", broadcasterID)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		set := FfzEmotes{}
		if err := json.Unmarshal(body, &set); err != nil {
			fmt.Printf("\t %s not registered for Ffz.\n", user)
		}

		count := 0

		for _, emotes := range set.Sets {
			for _, emote := range emotes.Emoticons {
				// global.ThirdPartyChannelEmotes[user] = append(global.ThirdPartyChannelEmotes[user], emote.Name)
				thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], emote.Name)
				count += 1
			}
		}

		//fmt.Printf("%s has %d FFZ emotes.", user, count)
	}
}

func GetLiveStatus(channelName string) (live bool) {
	url := "https://api.twitch.tv/helix/streams?user_login=" + channelName
	var jsonStr = []byte(`{"":""}`)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer "+global.TwitchAccessToken)
	req.Header.Set("Client-Id", global.TwitchClientID)
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var stream StreamStatusData
	if err := json.Unmarshal(body, &stream); err != nil {
		fmt.Println(err)
		return
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

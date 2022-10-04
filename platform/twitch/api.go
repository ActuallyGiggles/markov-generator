package twitch

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"markov-generator/global"
	"markov-generator/terminal"
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
	temp := make(map[string]Data)

	for _, channel := range global.Directives {
		data, ok := GetBroadcasterInfo(channel.ChannelName)
		if ok {
			temp[channel.ChannelName] = data
		}
	}

	Broadcasters = make(map[string]Data)
	Broadcasters = temp
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
		e := global.Emote{
			Name: emote.Name,
			Url:  emote.Images["url_4x"],
		}

		global.GlobalEmotes = append(global.GlobalEmotes, e)
	}
}

func getTwitchChannelEmotes() {
	for _, d := range Broadcasters {
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
			e := global.Emote{
				Name: emote.Name,
				Url:  emote.Images["url_4x"],
			}

			global.GlobalEmotes = append(global.GlobalEmotes, e)
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
		e := global.Emote{
			Name: emote.Name,
			Url:  emote.Urls[3][1],
		}

		global.GlobalEmotes = append(global.GlobalEmotes, e)
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
			e := global.Emote{
				Name: emote.Name,
				Url:  emote.Urls[3][1],
			}
			thirdPartyChannelEmotesToUpdate[channel.ChannelName] = append(thirdPartyChannelEmotesToUpdate[channel.ChannelName], e)
			count += 1
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
		e := global.Emote{
			Name: emote.Name,
			Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
		}

		global.GlobalEmotes = append(global.GlobalEmotes, e)
	}
}

func getBttvChannelEmotes() {
	for _, d := range Broadcasters {
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

		for _, emote := range emotes.ChannelEmotes {
			e := global.Emote{
				Name: emote.Name,
				Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
			}
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], e)
		}
		for _, emote := range emotes.SharedEmotes {
			e := global.Emote{
				Name: emote.Name,
				Url:  "https://cdn.betterttv.net/emote/" + emote.ID + "/3x.png",
			}
			thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], e)
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
	set := FfzSets{}
	if err := json.Unmarshal(body, &set); err != nil {
		panic(err)
	}

	for _, emotes := range set.Sets {
		for _, emote := range emotes.Emoticons {
			e := global.Emote{
				Name: emote.Name,
			}
			for size, url := range emote.Urls {
				switch size {
				case "4":
					e.Url = "https:" + url
					break
				case "2":
					e.Url = "https:" + url
					break
				case "1":
					e.Url = "https:" + url
					break
				}
			}
			global.GlobalEmotes = append(global.GlobalEmotes, e)
		}
	}
}

func getFfzChannelEmotes() {
	for _, d := range Broadcasters {
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

		set := FfzSets{}
		if err := json.Unmarshal(body, &set); err != nil {
			log.Printf("\t getFfzChannelEmotes failed\n")
			log.Printf("\t For channel %s\n1", ID)
			log.Println(err)
		}

		for _, emotes := range set.Sets {
			for _, emote := range emotes.Emoticons {
				e := global.Emote{
					Name: emote.Name,
				}
				for size, url := range emote.Urls {
					switch size {
					case "4":
						e.Url = "https:" + url
						break
					case "2":
						e.Url = "https:" + url
						break
					case "1":
						e.Url = "https:" + url
						break
					}
				}
				thirdPartyChannelEmotesToUpdate[user] = append(thirdPartyChannelEmotesToUpdate[user], e)
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

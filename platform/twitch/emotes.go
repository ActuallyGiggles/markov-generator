package twitch

import (
	"log"
	"markov-generator/global"
	"markov-generator/stats"
	"sync"
)

var (
	Broadcasters   map[string]Data
	broadcastersMx sync.Mutex

	globalEmotesToUpdate []global.Emote

	twitchChannelEmotesToUpdate []global.Emote

	thirdPartyChannelEmotesToUpdate   map[string][]global.Emote
	thirdPartyChannelEmotesToUpdateMx sync.Mutex
)

func GetEmoteController(isInit bool, channel global.Directive) (ok bool) {
	broadcastersMx.Lock()
	thirdPartyChannelEmotesToUpdateMx.Lock()
	defer broadcastersMx.Unlock()
	defer thirdPartyChannelEmotesToUpdateMx.Unlock()
	Broadcasters = make(map[string]Data)
	globalEmotesToUpdate = nil
	thirdPartyChannelEmotesToUpdate = make(map[string][]global.Emote)
	Broadcasters = make(map[string]Data)

	if channel.ChannelName == "" {
		for _, directive := range global.Directives {
			routineBroadcastersUpdate(directive)
		}

		if isInit {
			getTwitchGlobalEmotes()
			get7tvGlobalEmotes()
			getBttvGlobalEmotes()
			getFfzGlobalEmotes()
		}

		for _, c := range Broadcasters {
			getTwitchChannelEmotes(c)
			get7tvChannelEmotes(c)
			getBttvChannelEmotes(c)
			getFfzChannelEmotes(c)
		}

		transferEmotes(isInit)
	} else {
		// Get Broadcaster Info
		data, err := GetBroadcasterInfo(channel.ChannelName)
		if err != nil {
			stats.Log(err.Error())
			return false
		}
		Broadcasters[channel.ChannelName] = data // Add broadcaster

		// Get Twitch Channel Emotes
		err = getTwitchChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
		}
		for _, emote := range twitchChannelEmotesToUpdate {
			global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, emote) // Add each twitch channel emote
		}

		// Get 7tv emotes
		err = get7tvChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
			return false
		}

		// Get BTTV emotes
		err = getBttvChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
			return false
		}

		// Get FFZ emotes
		err = getFfzChannelEmotes(data)
		if err != nil {
			stats.Log(err.Error())
			return false
		}

		// Add each 7tv, BTTV, FFZ emote
		global.ThirdPartyChannelEmotes[channel.ChannelName] = thirdPartyChannelEmotesToUpdate[channel.ChannelName]
	}

	return true
}

func transferEmotes(isInit bool) {
	global.EmotesMx.Lock()
	defer global.EmotesMx.Unlock()

	if isInit {
		transferGlobalEmotes()
	}

	transferTwitchChannelEmotes()
	transferThirdPartyEmotes()
}

func transferGlobalEmotes() {
	global.GlobalEmotes = nil
	total := 0
	for _, emote := range globalEmotesToUpdate {
		global.GlobalEmotes = append(global.GlobalEmotes, emote)
		total++
	}
	log.Printf("[Updated %d Global emotes]", total)
}

func transferTwitchChannelEmotes() {
	global.TwitchChannelEmotes = nil
	total := 0
	for _, emote := range twitchChannelEmotesToUpdate {
		global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, emote)
		total++
	}
	log.Printf("[Updated %d Twitch Channel emotes]", total)
}

func transferThirdPartyEmotes() {
	global.ThirdPartyChannelEmotes = make(map[string][]global.Emote)
	total := 0
	for channelName, emoteSlice := range thirdPartyChannelEmotesToUpdate {
		global.ThirdPartyChannelEmotes[channelName] = emoteSlice
		total += len(emoteSlice)
	}
	log.Printf("[Updated %d Third Party emotes]", total)
}

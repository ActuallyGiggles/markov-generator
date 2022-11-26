package twitch

import (
	"log"
	"markov-generator/global"
	"markov-generator/stats"
	"sync"
)

var (
	Broadcasters   = make(map[string]Data)
	broadcastersMx sync.Mutex

	globalEmotesToUpdate              []global.Emote
	twitchChannelEmotesToUpdate       []global.Emote
	thirdPartyChannelEmotesToUpdate   []global.ThirdPartyEmotes
	thirdPartyChannelEmotesToUpdateMx sync.Mutex
)

func GetEmoteController(isInit bool, channel global.Directive) (ok bool) {
	broadcastersMx.Lock()
	thirdPartyChannelEmotesToUpdateMx.Lock()
	defer broadcastersMx.Unlock()
	defer thirdPartyChannelEmotesToUpdateMx.Unlock()
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
		twitchChannelEmotesToUpdate = nil

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
		e := global.ThirdPartyEmotes{
			Name:   channel.ChannelName,
			Emotes: thirdPartyChannelEmotesToUpdate[0].Emotes,
		}
		global.ThirdPartyChannelEmotes = append(global.ThirdPartyChannelEmotes, e)
		thirdPartyChannelEmotesToUpdate = nil
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
	globalEmotesToUpdate = nil
	log.Printf("[Updated %d Global emotes]", total)
}

func transferTwitchChannelEmotes() {
	global.TwitchChannelEmotes = nil
	total := 0
	for _, emote := range twitchChannelEmotesToUpdate {
		global.TwitchChannelEmotes = append(global.TwitchChannelEmotes, emote)
		total++
	}
	twitchChannelEmotesToUpdate = nil
	log.Printf("[Updated %d Twitch Channel emotes]", total)
}

func transferThirdPartyEmotes() {
	global.ThirdPartyChannelEmotes = nil
	total := 0
	for _, channel := range thirdPartyChannelEmotesToUpdate {
		e := global.ThirdPartyEmotes{
			Name:   channel.Name,
			Emotes: channel.Emotes,
		}

		global.ThirdPartyChannelEmotes = append(global.ThirdPartyChannelEmotes, e)
		total += len(channel.Emotes)
	}
	thirdPartyChannelEmotesToUpdate = nil
	log.Printf("[Updated %d Third Party emotes]", total)
}

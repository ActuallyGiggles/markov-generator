package twitch

import (
	"log"
	"markov-generator/global"
	"sync"
)

var (
	Broadcasters                      map[string]Data
	broadcastersMx                    sync.Mutex
	thirdPartyChannelEmotesToUpdate   map[string][]global.Emote
	thirdPartyChannelEmotesToUpdateMx sync.Mutex
)

func GetEmoteController(isInit bool) (ok bool) {
	broadcastersMx.Lock()
	thirdPartyChannelEmotesToUpdateMx.Lock()
	defer broadcastersMx.Unlock()
	defer thirdPartyChannelEmotesToUpdateMx.Unlock()
	Broadcasters = make(map[string]Data)
	thirdPartyChannelEmotesToUpdate = make(map[string][]global.Emote)

	if isInit {
		global.ChannelEmotesMx.Lock()
		defer global.ChannelEmotesMx.Unlock()
		getBroadcasterIDs()
		getTwitchGlobalEmotes()
		getTwitchChannelEmotes()
		get7tvGlobalEmotes()
		get7tvChannelEmotes()
		getBttvGlobalEmotes()
		getBttvChannelEmotes()
		getFfzGlobalEmotes()
		getFfzChannelEmotes()
		cleanAndTransferChannelEmotes()
	} else {
		err := getBroadcasterIDs()
		if err != nil {
			log.Println(err)
			return false
		}
		err = get7tvChannelEmotes()
		if err != nil {
			log.Println(err)
			return false
		}
		err = getBttvChannelEmotes()
		if err != nil {
			log.Println(err)
			return false
		}
		err = getFfzChannelEmotes()
		if err != nil {
			log.Println(err)
			return false
		}
		global.ChannelEmotesMx.Lock()
		defer global.ChannelEmotesMx.Unlock()
		cleanAndTransferChannelEmotes()
	}

	return true
}

func cleanAndTransferChannelEmotes() {
	global.ThirdPartyChannelEmotes = make(map[string][]global.Emote)
	for channelName, emoteSlice := range thirdPartyChannelEmotesToUpdate {
		global.ThirdPartyChannelEmotes[channelName] = emoteSlice
	}
	thirdPartyChannelEmotesToUpdate = make(map[string][]global.Emote)
}

package twitch

import (
	"MarkovGenerator/global"
)

var (
	broadcasterIDs                  map[string]string
	thirdPartyChannelEmotesToUpdate map[string][]string
)

func getEmoteController() {
	broadcasterIDs = make(map[string]string)
	thirdPartyChannelEmotesToUpdate = make(map[string][]string)

	if !didInitializationHappen {
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
		getBroadcasterIDs()
		get7tvChannelEmotes()
		getBttvChannelEmotes()
		getFfzChannelEmotes()
		global.ChannelEmotesMx.Lock()
		defer global.ChannelEmotesMx.Unlock()
		cleanAndTransferChannelEmotes()
	}

	didInitializationHappen = true
}

func cleanAndTransferChannelEmotes() {
	global.ThirdPartyChannelEmotes = make(map[string][]string)
	for channelName, emoteSlice := range thirdPartyChannelEmotesToUpdate {
		global.ThirdPartyChannelEmotes[channelName] = emoteSlice
	}
	thirdPartyChannelEmotesToUpdate = make(map[string][]string)
}

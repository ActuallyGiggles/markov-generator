package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/terminal"
	"sync"
)

var (
	Broadcasters                      map[string]Data
	broadcastersMx                    sync.Mutex
	thirdPartyChannelEmotesToUpdate   map[string][]global.Emote
	thirdPartyChannelEmotesToUpdateMx sync.Mutex
)

func GetEmoteController() {
	broadcastersMx.Lock()
	thirdPartyChannelEmotesToUpdateMx.Lock()
	defer broadcastersMx.Unlock()
	defer thirdPartyChannelEmotesToUpdateMx.Unlock()
	Broadcasters = make(map[string]Data)
	thirdPartyChannelEmotesToUpdate = make(map[string][]global.Emote)

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

	terminal.UpdateTerminal("emotes")
}

func cleanAndTransferChannelEmotes() {
	global.ThirdPartyChannelEmotes = make(map[string][]global.Emote)
	for channelName, emoteSlice := range thirdPartyChannelEmotesToUpdate {
		global.ThirdPartyChannelEmotes[channelName] = emoteSlice
	}
	thirdPartyChannelEmotesToUpdate = make(map[string][]global.Emote)
}

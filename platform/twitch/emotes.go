package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/terminal"
)

var (
	Broadcasters                    map[string]Data
	thirdPartyChannelEmotesToUpdate map[string][]string
)

func GetEmoteController() {
	Broadcasters = make(map[string]Data)
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
	global.ThirdPartyChannelEmotes = make(map[string][]string)
	for channelName, emoteSlice := range thirdPartyChannelEmotesToUpdate {
		global.ThirdPartyChannelEmotes[channelName] = emoteSlice
	}
	thirdPartyChannelEmotesToUpdate = make(map[string][]string)
}

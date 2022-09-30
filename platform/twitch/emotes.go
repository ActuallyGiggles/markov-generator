package twitch

import (
	"MarkovGenerator/global"
	"MarkovGenerator/terminal"
)

var (
	Broadcasters                    map[string]Data
	thirdPartyChannelEmotesToUpdate map[string][]global.Emote
)

func GetEmoteController() {
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

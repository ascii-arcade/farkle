package message

type Channel string

const (
	ChannelPing   Channel = "ping"
	ChannelLobby  Channel = "lobby"
	ChannelPlayer Channel = "player"
	ChannelGame   Channel = "game"
)

package pushclient

type Response map[string]interface{}

type HandshakeMessage struct {
	MessageType string `json:"messageType"`
	Uaid string `json:"uaid"`
	ChannelIDs []string `json:"channelIDs"`
}

type RegisterMessage struct {
	MessageType string `json:"messageType"`
	ChannelID string `json:"channelID"`
}

type AckMessage struct {
	MessageType string `json:"messageType"`
	Updates []*Update `json:"updates"`
}

type RegisterResponse struct {
	ChannelID string
	Status int
	PushEndpoint string
}

type Notification struct {
	Updates []*Update
}

type Update struct {
	ChannelID string `json:"channelID"`
	Version int `json:"version"`
}

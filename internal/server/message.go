package server

import (
	"encoding/json"
	"time"
)

type Message struct {
	Channel channel   `json:"channel"`
	Data    []byte    `json:"data"`
	SentAt  time.Time `json:"sent_at"`
}

func (m *Message) toBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

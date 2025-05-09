package server

import (
	"encoding/json"
	"time"
)

type Message struct {
	Channel Channel     `json:"channel"`
	Type    MessageType `json:"type"`
	Data    any         `json:"data"`
	SentAt  time.Time   `json:"sent_at"`

	from *client `json:"-"`
}

func (m *Message) toBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Channel Channel     `json:"channel"`
	Type    MessageType `json:"type"`
	Data    any         `json:"data"`
	SentAt  time.Time   `json:"sent_at"`

	from string `json:"-"`
}

func (m *Message) ToBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func (m *Message) IsFromPlayer(id string) bool {
	return m.from == id
}

package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Channel  Channel     `json:"channel"`
	Type     MessageType `json:"type"`
	Data     string      `json:"data"`
	SentAt   time.Time   `json:"sent_at"`
	PlayerId string      `json:"player_id"`
}

func (m *Message) ToBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func (m *Message) Unmarshal(v any) error {
	return json.Unmarshal([]byte(m.Data), v)
}

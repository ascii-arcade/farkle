package server

import (
	"encoding/json"
	"time"
)

type message struct {
	channel channel   `json:"channel"`
	data    []byte    `json:"data"`
	sentAt  time.Time `json:"sent_at"`
}

func toMessage(b []byte) (*message, error) {
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *message) toBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

package player

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
)

func (p *Player) MonitorMessages(lobbyChan chan message.Message) {
	if p.conn == nil {
		return
	}
	go func() {
		for {
			msg, err := p.ReceiveMessage()
			if err != nil {
				break
			}

			if msg.Channel == message.ChannelPing {
				p.LastSeen = time.Now()
				continue
			}

			msg.PlayerId = p.Id

			lobbyChan <- msg
		}
	}()
}

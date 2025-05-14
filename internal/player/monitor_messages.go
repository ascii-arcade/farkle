package player

import (
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
)

func (p *Player) MonitorMessages(messageChan chan message.Message) (errOut error) {
	if p.conn == nil {
		return nil
	}
	go func() {
		for {
			msg, err := p.ReceiveMessage()
			if err != nil {
				errOut = err
				return
			}

			switch msg.Channel {
			case message.ChannelPing:
				p.lastSeen = time.Now()
				continue
			case message.ChannelPlayer:
				switch msg.Type {
				case message.MessageTypeMe:
					pIn := Player{}
					if err := msg.Unmarshal(&pIn); err != nil {
						p.logger.Error("error unmarshalling player message", "error", err)
						continue
					}
					p.Update(pIn)
				}
			}

			msg.PlayerId = p.Id

			messageChan <- msg
		}
	}()

	return errOut
}

// package player

// import (
// 	"encoding/json"
// 	"time"

// 	"golang.org/x/net/websocket"
// )

// type Player struct {
// 	Name   string `json:"name"`
// 	Score  int    `json:"score"`
// 	Host   bool   `json:"host"`
// }

// func New(name string) *Player {
// 	return &Player{
// 		Name:  name,
// 		Score: 0,
// 	}
// }

// func (p *Player) ToBytes() []byte {
// 	b, err := json.Marshal(p)
// 	if err != nil {
// 		return nil
// 	}
// 	return b
// }

//	func FromMap(m map[string]interface{}) (*Player, error) {
//		b, err := json.Marshal(m)
//		if err != nil {
//			return nil, err
//		}
//		p := &Player{}
//		if err := json.Unmarshal(b, p); err != nil {
//			return nil, err
//		}
//		return p, nil
//	}
package player

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ascii-arcade/farkle/internal/message"
	"github.com/rs/xid"
	"golang.org/x/net/websocket"
)

type Player struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Host   bool   `json:"host"`
	Active bool   `json:"active"`

	LastSeen time.Time       `json:"-"`
	conn     *websocket.Conn `json:"-"`
	// player   *player.Player `json:"-"`
}

func NewPlayer(conn *websocket.Conn, name string) *Player {
	c := &Player{
		Id:       xid.New().String(),
		Name:     name,
		Active:   true,
		LastSeen: time.Now().Add(15 * time.Second),

		conn: conn,
	}
	return c
}

func (p *Player) ToJSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *Player) SendMessage(msg message.Message) error {
	if p.conn == nil {
		return nil
	}
	if err := websocket.JSON.Send(p.conn, msg); err != nil {
		return err
	}
	return nil
}

func (p *Player) ReceiveMessage() (message.Message, error) {
	var msg message.Message
	if p.conn == nil {
		return message.Message{}, errors.New("no connection")
	}
	if err := websocket.JSON.Receive(p.conn, &msg); err != nil {
		return message.Message{}, err
	}
	return msg, nil
}

func (p *Player) Close() error {
	if p.conn == nil {
		return nil
	}
	if err := p.conn.Close(); err != nil {
		return err
	}
	p.Active = false
	return nil
}

func (p *Player) Connected() bool {
	return p.conn != nil
}

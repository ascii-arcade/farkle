package eventloop

import (
	"github.com/ascii-arcade/farkle/internal/message"
	tea "github.com/charmbracelet/bubbletea"
)

type EventLoop struct {
	Incoming   chan message.Message // NetworkManager's message channel
	DispatchTo tea.Program          // Bubble Tea instance to send messages
}

func New(incoming chan message.Message, program *tea.Program) *EventLoop {
	return &EventLoop{
		Incoming:   incoming,
		DispatchTo: *program,
	}
}

func (el *EventLoop) Start() {
	go func() {
		for msg := range el.Incoming {
			el.DispatchTo.Send(NetworkMsg{Data: msg})
		}
	}()
}

type NetworkMsg struct {
	Data message.Message
}

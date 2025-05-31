package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		conn, _, err := websocket.DefaultDialer.Dial(m.url, nil)
		return connectMsg{conn, err}
	}
}

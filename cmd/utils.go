package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func wrapText(text string, width int) []string {
	var lines []string
	for len(text) > width {
		idx := strings.LastIndex(text[:width+1], " ")
		if idx <= 0 {
			idx = width
		}
		lines = append(lines, text[:idx])
		text = strings.TrimSpace(text[idx:])
	}
	lines = append(lines, text)
	return lines
}

func readWsCmd(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return errMsg{err}
		}
		return wsMsg{string(msg)}
	}
}

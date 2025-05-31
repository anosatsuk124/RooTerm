package main

import (
	"encoding/json"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectMsg:
		if msg.err != nil {
			m.err = msg.err
			m.msgs = append(m.msgs, "Connection error: "+msg.err.Error())
			return m, nil
		}
		m.conn = msg.conn
		m.msgs = append(m.msgs, "Connected to "+m.url)
		return m, readWsCmd(m.conn)

	case wsMsg:
		var cp ChatPayload
		if err := json.Unmarshal([]byte(msg.text), &cp); err == nil {
			if cp.IsReasoning {
				m.msgs = append(m.msgs, "\033[90m"+cp.Message+"\033[0m")
			} else {
				m.msgs = append(m.msgs, cp.Message)
			}
			return m, readWsCmd(m.conn)
		}
		m.msgs = append(m.msgs, msg.text)
		return m, readWsCmd(m.conn)

	case errMsg:
		m.msgs = append(m.msgs, "Read error: "+msg.err.Error())
		return m, nil

	case tea.KeyMsg:
		if m.showSuggest {
			switch msg.Type {
			case tea.KeyEnter:
				choice := m.suggestList.SelectedItem().(suggestItem)
				p := Payload{Message: string(choice), Time: time.Now().Format(time.RFC3339)}
				b, err := json.Marshal(p)
				if err != nil {
					m.msgs = append(m.msgs, "JSON marshal error: "+err.Error())
				} else {
					_ = m.conn.WriteMessage(websocket.TextMessage, b)
				}
				m.showSuggest = false
				m.textInput.Focus()
			default:
				var cmd tea.Cmd
				m.suggestList, cmd = m.suggestList.Update(msg)
				return m, cmd
			}
			return m, readWsCmd(m.conn)
		}
		// 通常入力モード
		switch msg.Type {
		case tea.KeyEnter:
			p := Payload{Message: m.textInput.Value(), Time: time.Now().Format(time.RFC3339)}
			b, err := json.Marshal(p)
			if err != nil {
				m.msgs = append(m.msgs, "JSON marshal error: "+err.Error())
			} else {
				_ = m.conn.WriteMessage(websocket.TextMessage, b)
			}
			m.textInput.SetValue("")
			return m, nil
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

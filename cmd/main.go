package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

// Payload represents the JSON payload to send.
type Payload struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

// SuggestionPayload is the structure with question + suggest list.
type SuggestionPayload struct {
	Question string   `json:"question"`
	Suggest  []string `json:"suggest"`
}

type ChatPayload struct {
	Message     string `json:"message"`
	IsReasoning bool   `json:"is_reasoning"`
}

// Model is our Bubble Tea model.
type Model struct {
	url       string
	conn      *websocket.Conn
	textInput textinput.Model
	msgs      []string
	err       error
	// suggestion 用
	showSuggest bool
	suggestList list.Model
}

// メッセージ用の list.Item 実装
type suggestItem string

func (i suggestItem) Title() string       { return string(i) }
func (i suggestItem) Description() string { return "" }
func (i suggestItem) FilterValue() string { return string(i) }

type connectMsg struct {
	conn *websocket.Conn
	err  error
}

type wsMsg struct {
	text string
}

type errMsg struct {
	err error
}

func main() {
	host := flag.String("host", "127.0.0.1", "Target host")
	port := flag.Int("port", 9421, "Target port")
	flag.Parse()
	url := fmt.Sprintf("ws://%s:%d/", *host, *port)

	// textinput 初期化
	ti := textinput.New()
	ti.Placeholder = "Type a message"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	m := Model{
		url:       url,
		textInput: ti,
		msgs:      []string{},
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		conn, _, err := websocket.DefaultDialer.Dial(m.url, nil)
		return connectMsg{conn, err}
	}
}

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
		// Reasoning の場合はメッセージを灰色にする
		// それ以外は通常のメッセージとして扱う
		var cp ChatPayload
		if err := json.Unmarshal([]byte(msg.text), &cp); err == nil {
			if cp.IsReasoning {
				// Reasoning メッセージ
				m.msgs = append(m.msgs, "\033[90m"+cp.Message+"\033[0m") // ANSI escape code for gray
				return m, readWsCmd(m.conn)
			}

			m.msgs = append(m.msgs, cp.Message)
			return m, readWsCmd(m.conn)
		}
		// 受信メッセージが JSON でない場合はそのまま表示
		m.msgs = append(m.msgs, msg.text)
		return m, readWsCmd(m.conn)

	case errMsg:
		m.msgs = append(m.msgs, "Read error: "+msg.err.Error())
		return m, nil

	case tea.KeyMsg:
		// 選択肢表示中なら list で操作
		if m.showSuggest {
			switch msg.Type {
			case tea.KeyEnter:
				choice := m.suggestList.SelectedItem().(suggestItem)
				// 選択肢をそのまま送信
				payload := Payload{
					Message: string(choice),
					Time:    time.Now().Format(time.RFC3339),
				}
				b, err := json.Marshal(payload)
				if err != nil {
					m.msgs = append(m.msgs, "JSON marshal error: "+err.Error())
				} else {
					m.conn.WriteMessage(websocket.TextMessage, b)
				}
				// 選択肢モード終了して入力欄にフォーカス
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
			payload := Payload{
				Message: m.textInput.Value(),
				Time:    time.Now().Format(time.RFC3339),
			}
			b, err := json.Marshal(payload)
			if err != nil {
				m.msgs = append(m.msgs, "JSON marshal error: "+err.Error())
			} else {
				m.conn.WriteMessage(websocket.TextMessage, b)
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

// wrapText は文字列を指定幅で改行するヘルパー関数です。
func wrapText(text string, width int) []string {
	var lines []string
	for len(text) > width {
		// 幅を超える部分で最後のスペースを探す
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

func (m Model) View() string {
	var b strings.Builder
	for _, line := range m.msgs {
		for _, ln := range wrapText(line, 80) {
			b.WriteString(ln + "\n")
		}
	}

	if m.showSuggest {
		// 選択肢リストを描画
		b.WriteString("\n" + m.suggestList.View())
	} else {
		// テキスト入力欄を描画
		b.WriteString("\n" + m.textInput.View())
	}

	return b.String()
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

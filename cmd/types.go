package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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

// ChatPayload はサーバーから来るチャットメッセージ構造体。
type ChatPayload struct {
	Message     string `json:"message"`
	IsReasoning bool   `json:"is_reasoning"`
}

// Bubble Tea 用モデル
type Model struct {
	url         string
	conn        *websocket.Conn
	textInput   textinput.Model
	msgs        []string
	err         error
	showSuggest bool
	suggestList list.Model
}

// list.Item 実装
type suggestItem string

func (i suggestItem) Title() string       { return string(i) }
func (i suggestItem) Description() string { return "" }
func (i suggestItem) FilterValue() string { return string(i) }

// 内部メッセージ型
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

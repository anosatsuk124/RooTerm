package main

import (
	"flag"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

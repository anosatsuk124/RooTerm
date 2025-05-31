package main

import "strings"

func (m Model) View() string {
	var b strings.Builder
	for _, line := range m.msgs {
		for _, ln := range wrapText(line, 80) {
			b.WriteString(ln + "\n")
		}
	}
	if m.showSuggest {
		b.WriteString("\n" + m.suggestList.View())
	} else {
		b.WriteString("\n" + m.textInput.View())
	}
	return b.String()
}

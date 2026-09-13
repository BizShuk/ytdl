package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the interactive monitor TUI for the specified directory.
func Run(dir string) error {
	p := tea.NewProgram(NewModel(dir), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

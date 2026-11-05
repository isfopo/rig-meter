// Package ui holds the app-specific terminal UI for this template.
//
// Replace this minimal model with your application's Bubble Tea model.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is a tiny placeholder TUI used to prove the template builds.
type Model struct{}

// New returns the template model.
func New() Model { return Model{} }

func (Model) Init() tea.Cmd { return nil }

func (Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return Model{}, tea.Quit
		}
	}
	return Model{}, nil
}

func (Model) View() string {
	return "rig template\n\nPress q to quit.\n"
}

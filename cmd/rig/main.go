// Command rig-meter is a standalone capture-level meter.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/isfopo/rig-meter/internal/ui"
)

func main() {
	program := tea.NewProgram(ui.New(), tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "rig-meter:", err)
		os.Exit(1)
	}
	if m, ok := final.(ui.Model); ok {
		m.Shutdown()
	}
}

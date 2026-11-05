// Command rig is a minimal template for a standalone rig application.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/isfopo/rig/internal/ui"
)

func main() {
	program := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "rig:", err)
		os.Exit(1)
	}
}

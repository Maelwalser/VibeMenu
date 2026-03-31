package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vibe-mvp/internal/manifest"
	"github.com/vibe-mvp/internal/ui"
)

const manifestPath = "manifest.json"

func main() {
	saveFunc := func(m *manifest.Manifest) error {
		return m.Save(manifestPath)
	}

	model := ui.NewModel(saveFunc)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// After the TUI exits, print a summary if the manifest was populated.
	if m, ok := finalModel.(ui.Model); ok {
		mf := m.BuildManifest()
		if mf.Backend.ArchPattern != "" {
			fmt.Printf("\nManifest saved to %s\n", manifestPath)
			fmt.Printf("Backend   : %s  [%s]\n", mf.Backend.ArchPattern, mf.Backend.ComputeEnv)
			fmt.Printf("Entities  : %d defined\n", len(mf.Entities))
			fmt.Printf("Databases : %d defined\n", len(mf.Databases))
			fmt.Printf("Services  : %d defined\n", len(mf.Backend.Services))
		}
	}
}

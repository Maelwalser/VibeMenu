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

	// After the TUI exits, check if there's a completed manifest to report.
	if m, ok := finalModel.(ui.Model); ok {
		mf := m.BuildManifest()
		if mf.Architecture.TargetEnvironment != "" {
			fmt.Printf("\nManifest saved to %s\n", manifestPath)
			fmt.Printf("Architecture : %s / %s\n",
				mf.Architecture.TargetEnvironment,
				mf.Architecture.Topology)
			fmt.Printf("Backend      : %s %s\n",
				mf.TechStack.BackendLanguage,
				mf.TechStack.BackendFramework)
		}
	}
}

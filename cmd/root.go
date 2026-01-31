package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/libcat/libcat/internal/storage"
	"github.com/libcat/libcat/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "libcat",
	Short: "A TUI book tracking application",
	Long: `LibCat is a terminal-based book tracking application.

Track your book library, reading progress, and discover new books
using the Open Library API.

Run without arguments to start the TUI, or use subcommands for quick actions.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runTUI() {
	store, err := storage.NewStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	library, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading library: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(library, store)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

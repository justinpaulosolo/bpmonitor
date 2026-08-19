package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/justinpaulosolo/bpmonitor/internal/storage"
	"github.com/justinpaulosolo/bpmonitor/internal/tui"
)

func main() {
	store, err := storage.Open("bpmonitor.db")
	if err != nil {
		fmt.Println("failed to open storage:", err)
		os.Exit(1)
	}
	defer store.Close()

	if _, err := tea.NewProgram(tui.NewModel(store)).Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kingshuk7995/pktfs/client"
)

func main() {
	addr := "localhost:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	api, err := client.ConnectAPI(addr)
	if err != nil {
		log.Fatal("Could not connect to server:", err)
	}
	defer api.Close()

	m, err := client.NewTUI(api)
	if err != nil {
		log.Fatal("Could not initialize TUI:", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

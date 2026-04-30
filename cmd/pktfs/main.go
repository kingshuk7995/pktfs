package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kingshuk7995/pktfs/client"
	"github.com/kingshuk7995/pktfs/server"

	tea "github.com/charmbracelet/bubbletea"
)

func usage() {
	fmt.Fprintf(os.Stderr, `pktfs

Usage:
  pktfs --serve --port 8080 --root ./data
  pktfs --connect <host> <port>

Flags:
`)
	flag.PrintDefaults()
}

func main() {
	serveMode := flag.Bool("serve", false, "run as server")
	connectMode := flag.Bool("connect", false, "run as client")
	port := flag.String("port", "8080", "server port")
	root := flag.String("root", "./data", "server root directory")
	flag.Usage = usage
	flag.Parse()

	switch {
	case *serveMode && *connectMode:
		log.Fatal("choose only one mode: --serve or --connect")

	case *serveMode:
		if err := os.MkdirAll(*root, 0o755); err != nil {
			log.Fatal(err)
		}

		srv, err := server.New(":"+*port, *root)
		if err != nil {
			log.Fatal(err)
		}

		log.Fatal(srv.ListenAndServe())

	case *connectMode:
		if flag.NArg() != 2 {
			usage()
			os.Exit(2)
		}

		addr := flag.Arg(0) + ":" + flag.Arg(1)

		api, err := client.ConnectAPI(addr)
		if err != nil {
			log.Fatal("could not connect to server:", err)
		}
		defer api.Close()

		m := client.NewShell(api)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

	default:
		usage()
		os.Exit(2)
	}
}

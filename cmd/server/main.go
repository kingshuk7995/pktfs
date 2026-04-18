package server

import (
	"log"
	"net"
	"os"
	"github.com/kingshuk7995/pktfs/server"
)

func main() {
	os.MkdirAll("./data", 0755)
	s := server.New(":8080", "./data")

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Server listening on", s.Addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go s.Handle(conn)
	}
}

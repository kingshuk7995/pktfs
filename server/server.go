package server

import (
	"github.com/kingshuk7995/pktfs/utils"
	"net"
	"path/filepath"
)

type Server struct {
	Addr string
	root string
	lm   *utils.LockManager
}

func New(addr, root string) *Server {
	root = filepath.Clean(root)
	return &Server{
		Addr: addr,
		root: root,
		lm:   utils.NewLockManager(root),
	}
}

func (s *Server) Handle(conn net.Conn) {
	session := NewSession(conn, s.root, s.lm)
	session.Run()
}

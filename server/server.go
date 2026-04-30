package server

import (
	"net"
	"path/filepath"
)

type Server struct {
	Addr string
	root string
	lm   *LockManager
}

func New(addr, root string) (*Server, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &Server{
		Addr: addr,
		root: absRoot,
		lm:   NewLockManager(absRoot),
	}, nil
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	session := NewSession(conn, s.root, s.lm)
	session.Run()
}

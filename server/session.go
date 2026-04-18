package server

import (
	"bufio"
	"net"
	"strings"
	"github.com/kingshuk7995/pktfs/utils"
)

type Session struct {
	conn   net.Conn
	reader *bufio.Reader
	cwd    string
	root   string
	lm     *utils.LockManager
}

func NewSession(conn net.Conn, root string, lm *utils.LockManager) *Session {
	return &Session{
		conn:   conn,
		reader: bufio.NewReader(conn),
		cwd:    root,
		root:   root,
		lm:     lm,
	}
}

func (s *Session) Run() {
	defer s.conn.Close()

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "PWD":
			s.handlePWD()

		case "CD":
			s.handleCD(parts)

		case "LIST":
			s.handleLIST()

		case "GET":
			s.handleGET(parts)

		case "PUT":
			s.handlePUT(parts)

		case "QUIT":
			s.writeOK()
			return

		default:
			s.writeErr("unknown command")
		}
	}
}

func (s *Session) writeOK() {
	s.conn.Write([]byte("OK\n"))
}

func (s *Session) writeErr(msg string) {
	s.conn.Write([]byte("ERR " + msg + "\n"))
}

func (s *Session) writeLine(msg string) {
	s.conn.Write([]byte(msg + "\n"))
}

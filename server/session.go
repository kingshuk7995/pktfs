package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Session struct {
	conn   net.Conn
	reader *bufio.Reader
	cwd    string
	root   string
	lm     *LockManager
}

func NewSession(conn net.Conn, root string, lm *LockManager) *Session {
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

func (s *Session) resolve(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidPath
	}

	var target string
	if filepath.IsAbs(path) {
		target = filepath.Clean(path)
	} else {
		target = filepath.Clean(filepath.Join(s.cwd, path))
	}

	rel, err := filepath.Rel(s.root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}

	return target, nil
}

func (s *Session) writeOK() {
	_, _ = s.conn.Write([]byte("OK\n"))
}

func (s *Session) writeOKLine(msg string) {
	_, _ = s.conn.Write([]byte("OK " + msg + "\n"))
}

func (s *Session) writeErr(msg string) {
	_, _ = s.conn.Write([]byte("ERR " + msg + "\n"))
}

func (s *Session) writeLine(msg string) {
	_, _ = s.conn.Write([]byte(msg + "\n"))
}

func (s *Session) handlePWD() {
	s.writeOKLine(s.cwd)
}

func (s *Session) handleCD(parts []string) {
	if len(parts) < 2 {
		s.writeErr("missing path")
		return
	}

	newPath, err := s.resolve(parts[1])
	if err != nil {
		s.writeErr("invalid directory")
		return
	}

	info, err := os.Stat(newPath)
	if err != nil || !info.IsDir() {
		s.writeErr("invalid directory")
		return
	}

	s.cwd = newPath
	s.writeOK()
}

func (s *Session) handleLIST() {
	files, err := os.ReadDir(s.cwd)
	if err != nil {
		s.writeErr("cannot list")
		return
	}

	for _, f := range files {
		info, _ := f.Info()
		if f.IsDir() {
			s.writeLine(fmt.Sprintf("DIR %s", f.Name()))
		} else {
			s.writeLine(fmt.Sprintf("FILE %s %d", f.Name(), info.Size()))
		}
	}

	s.writeLine("END")
}

func (s *Session) handleGET(parts []string) {
	if len(parts) < 2 {
		s.writeErr("missing file")
		return
	}

	full, err := s.resolve(parts[1])
	if err != nil {
		s.writeErr("invalid path")
		return
	}

	unlock, err := s.lm.RLock(full)
	if err != nil {
		s.writeErr("invalid path")
		return
	}
	defer unlock()

	f, err := os.Open(full)
	if err != nil {
		s.writeErr("cannot open")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		s.writeErr("cannot stat")
		return
	}

	s.writeOKLine(fmt.Sprintf("%d", info.Size()))
	_, _ = io.Copy(s.conn, f)
}

func (s *Session) handlePUT(parts []string) {
	if len(parts) < 3 {
		s.writeErr("usage PUT <file> <size>")
		return
	}

	full, err := s.resolve(parts[1])
	if err != nil {
		s.writeErr("invalid path")
		return
	}

	size, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || size < 0 {
		s.writeErr("invalid size")
		return
	}

	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		s.writeErr("cannot prepare directory")
		return
	}

	tmp := full + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		s.writeErr("cannot create")
		return
	}

	s.writeOK()

	_, copyErr := io.CopyN(f, s.reader, size)
	closeErr := f.Close()

	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmp)
		return
	}

	unlock, err := s.lm.Lock(full)
	if err != nil {
		_ = os.Remove(tmp)
		s.writeErr("lock error")
		return
	}
	defer unlock()

	if err := os.Rename(tmp, full); err != nil {
		_ = os.Remove(tmp)
		s.writeErr("rename failed")
		return
	}

	s.writeOK()
}

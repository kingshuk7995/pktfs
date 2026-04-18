package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

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

	path := parts[1]

	unlock, err := s.lm.RLock(s.cwd, path)
	if err != nil {
		s.writeErr("invalid path")
		return
	}
	defer unlock()

	full := s.absPath(path)

	f, err := os.Open(full)
	if err != nil {
		s.writeErr("cannot open")
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	s.writeLine(fmt.Sprintf("OK %d", info.Size()))

	io.Copy(s.conn, f)
}

func (s *Session) handlePUT(parts []string) {
	if len(parts) < 3 {
		s.writeErr("usage PUT <file> <size>")
		return
	}

	path := parts[1]
	size, err := strconv.Atoi(parts[2])
	if err != nil {
		s.writeErr("invalid size")
		return
	}

	full := s.absPath(path)
	tmp := full + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		s.writeErr("cannot create")
		return
	}

	s.writeOK()

	_, err = io.CopyN(f, s.reader, int64(size))
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return
	}

	unlock, err := s.lm.Lock(s.cwd, path)
	if err != nil {
		s.writeErr("lock error")
		return
	}
	defer unlock()

	os.Rename(tmp, full)

	s.writeOK()
}

func (s *Session) absPath(path string) string {
	return filepath.Clean(filepath.Join(s.cwd, path))
}

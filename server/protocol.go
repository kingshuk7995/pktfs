package server

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *Session) handlePWD() {
	s.writeLine(fmt.Sprintf("OK %s", s.cwd))
}

func (s *Session) handleCD(parts []string) {
	if len(parts) < 2 {
		s.writeErr("missing path")
		return
	}

	newPath := filepath.Join(s.cwd, parts[1])
	newPath = filepath.Clean(newPath)

	info, err := os.Stat(newPath)
	if err != nil || !info.IsDir() {
		s.writeErr("invalid directory")
		return
	}

	s.cwd = newPath
	s.writeOK()
}

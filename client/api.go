package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type API struct {
	conn   net.Conn
	reader *bufio.Reader
}

func ConnectAPI(addr string) (*API, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &API{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (a *API) Close() {
	if a.conn != nil {
		_ = a.conn.Close()
	}
}

func (a *API) send(cmd string) error {
	_, err := a.conn.Write([]byte(cmd + "\n"))
	return err
}

func (a *API) readLine() (string, error) {
	line, err := a.reader.ReadString('\n')
	return strings.TrimSpace(line), err
}

func (a *API) Pwd() (string, error) {
	if err := a.send("PWD"); err != nil {
		return "", err
	}

	line, err := a.readLine()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(line, "OK ") {
		return strings.TrimSpace(strings.TrimPrefix(line, "OK ")), nil
	}

	if line == "OK" {
		return "/", nil
	}

	return "", fmt.Errorf("error: %s", line)
}

type FileInfo struct {
	Name string
	Size int64
	Dir  bool
}

func (a *API) List() ([]FileInfo, error) {
	if err := a.send("LIST"); err != nil {
		return nil, err
	}

	var files []FileInfo
	for {
		line, err := a.readLine()
		if err != nil {
			return nil, err
		}

		if line == "END" {
			break
		}
		if strings.HasPrefix(line, "ERR") {
			return nil, fmt.Errorf("error: %s", line)
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "DIR":
			files = append(files, FileInfo{Name: parts[1], Dir: true})
		case "FILE":
			if len(parts) >= 3 {
				size, _ := strconv.ParseInt(parts[2], 10, 64)
				files = append(files, FileInfo{Name: parts[1], Size: size, Dir: false})
			}
		}
	}

	return files, nil
}

func (a *API) Cd(dir string) error {
	if err := a.send("CD " + dir); err != nil {
		return err
	}

	line, err := a.readLine()
	if err != nil {
		return err
	}

	if strings.HasPrefix(line, "OK") {
		return nil
	}

	return fmt.Errorf("error: %s", line)
}

func (a *API) Get(remote, local string) error {
	if err := a.send("GET " + remote); err != nil {
		return err
	}

	line, err := a.readLine()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(line, "OK") {
		return fmt.Errorf("error: %s", line)
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return fmt.Errorf("invalid response: %s", line)
	}

	size, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return err
	}

	file, err := os.Create(local)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.CopyN(file, a.reader, size)
	return err
}

func (a *API) Put(local, remote string) error {
	file, err := os.Open(local)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if err := a.send(fmt.Sprintf("PUT %s %d", remote, info.Size())); err != nil {
		return err
	}

	line, err := a.readLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "OK") {
		return fmt.Errorf("error: %s", line)
	}

	if _, err := io.Copy(a.conn, file); err != nil {
		return err
	}

	line, err = a.readLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "OK") {
		return fmt.Errorf("error: %s", line)
	}

	return nil
}

func (a *API) Quit() {
	_ = a.send("QUIT")
	a.Close()
}

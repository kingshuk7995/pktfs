package client

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type model struct {
	api   *API
	input textinput.Model

	lines []string
	cwd   string

	width  int
	height int
}

func NewShell(api *API) *model {
	ti := textinput.New()
	ti.Placeholder = "type help"
	ti.Prompt = "> "
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 80

	m := &model{
		api:   api,
		input: ti,
		lines: []string{},
		cwd:   "/",
	}

	if cwd, err := api.Pwd(); err == nil && cwd != "" {
		m.cwd = cwd
		m.lines = append(m.lines, "connected: "+cwd)
	} else if err != nil {
		m.lines = append(m.lines, "connected, but failed to read pwd: "+err.Error())
	}

	return m
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) appendLine(s string) {
	m.lines = append(m.lines, s)
	if len(m.lines) > 200 {
		m.lines = m.lines[len(m.lines)-200:]
	}
}

func (m *model) runCommand(cmd string) tea.Cmd {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	m.appendLine(m.input.Prompt + cmd)

	parts := strings.Fields(cmd)
	switch strings.ToLower(parts[0]) {
	case "help":
		m.appendLine("pwd | ls | cd <dir> | get <remote> [local] | put <local> [remote] | clear | quit")

	case "pwd":
		cwd, err := m.api.Pwd()
		if err != nil {
			m.appendLine("error: " + err.Error())
		} else {
			m.cwd = cwd
			m.appendLine(cwd)
		}

	case "ls":
		files, err := m.api.List()
		if err != nil {
			m.appendLine("error: " + err.Error())
			break
		}
		for _, f := range files {
			if f.Dir {
				m.appendLine("DIR  " + f.Name)
			} else {
				m.appendLine(fmt.Sprintf("FILE %s (%d bytes)", f.Name, f.Size))
			}
		}

	case "cd":
		if len(parts) < 2 {
			m.appendLine("error: missing path")
			break
		}
		if err := m.api.Cd(parts[1]); err != nil {
			m.appendLine("error: " + err.Error())
			break
		}
		cwd, err := m.api.Pwd()
		if err == nil {
			m.cwd = cwd
		}
		m.appendLine("cwd: " + m.cwd)

	case "get":
		if len(parts) < 2 {
			m.appendLine("error: missing remote file")
			break
		}
		remote := parts[1]
		local := filepath.Base(remote)
		if len(parts) >= 3 {
			local = parts[2]
		}
		if err := m.api.Get(remote, local); err != nil {
			m.appendLine("error: " + err.Error())
		} else {
			m.appendLine("downloaded: " + local)
		}

	case "put":
		if len(parts) < 2 {
			m.appendLine("error: missing local file")
			break
		}
		local := parts[1]
		remote := filepath.Base(local)
		if len(parts) >= 3 {
			remote = parts[2]
		}
		if err := m.api.Put(local, remote); err != nil {
			m.appendLine("error: " + err.Error())
		} else {
			m.appendLine("uploaded: " + remote)
		}

	case "clear":
		m.lines = nil

	case "quit", "exit":
		m.api.Quit()
		return tea.Quit

	default:
		m.appendLine("error: unknown command")
	}

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(20, msg.Width-8)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.api.Quit()
			return m, tea.Quit
		case tea.KeyEnter:
			cmd := m.input.Value()
			m.input.SetValue("")
			return m, m.runCommand(cmd)
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) View() string {
	header := titleStyle.Render("pktfs")
	if m.cwd != "" {
		header += " " + dimStyle.Render("["+m.cwd+"]")
	}

	body := strings.Join(m.tail(18), "\n")
	if body == "" {
		body = dimStyle.Render("type help")
	}

	view := fmt.Sprintf("%s\n\n%s\n\n%s",
		header,
		body,
		m.input.View(),
	)

	return frameStyle.Width(max(40, m.width-4)).Render(view)
}

func (m *model) tail(n int) []string {
	if len(m.lines) <= n {
		return m.lines
	}
	return m.lines[len(m.lines)-n:]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

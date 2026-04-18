package client

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1).MarginLeft(2)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).MarginTop(1).MarginLeft(2)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).MarginTop(1).MarginLeft(2)
)

type item struct {
	FileInfo
}

func (i item) Title() string {
	if i.Dir {
		return "📁 " + i.Name
	}
	return "📄 " + i.Name
}

func (i item) Description() string {
	if i.Dir {
		return "Directory"
	}
	return fmt.Sprintf("%d bytes", i.Size)
}

func (i item) FilterValue() string { return i.Name }

type model struct {
	api           *API
	list          list.Model
	textInput     textinput.Model
	cwd           string
	statusMessage string
	statusIsError bool
	inputMode     bool
}

func NewTUI(api *API) (*model, error) {
	m := &model{
		api: api,
	}

	ti := textinput.New()
	ti.Placeholder = "Enter local path to upload"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	m.textInput = ti

	err := m.refreshFiles()
	return m, err
}

func (m *model) refreshFiles() error {
	cwd, err := m.api.Pwd()
	if err != nil {
		return err
	}
	m.cwd = cwd

	files, err := m.api.List()
	if err != nil {
		return err
	}

	items := []list.Item{
		item{FileInfo{Name: "..", Dir: true}},
	}
	for _, f := range files {
		items = append(items, item{f})
	}

	delegate := list.NewDefaultDelegate()

	m.list = list.New(items, delegate, 50, 20)
	m.list.Title = "PKTFS: " + m.cwd

	return nil
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			switch msg.Type {
			case tea.KeyEnter:
				m.inputMode = false
				localObj := m.textInput.Value()
				if localObj != "" {
					info, err := os.Stat(localObj)
					if err != nil {
						m.setStatus("Error: "+err.Error(), true)
					} else if info.IsDir() {
						m.setStatus("Cannot upload directory", true)
					} else {
						remoteObj := info.Name()
						err = m.api.Put(localObj, remoteObj)
						if err != nil {
							m.setStatus("Upload failed: "+err.Error(), true)
						} else {
							m.setStatus("Uploaded "+remoteObj+" successfully!", false)
							m.refreshFiles()
						}
					}
				}
				return m, nil
			case tea.KeyEsc:
				m.inputMode = false
				return m, nil
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.api.Quit()
			return m, tea.Quit
		case "u":
			m.inputMode = true
			m.textInput.SetValue("")
			return m, nil
		case "r":
			if err := m.refreshFiles(); err != nil {
				m.setStatus("Refresh failed: "+err.Error(), true)
			} else {
				m.setStatus("Refreshed directory contents", false)
			}
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				if i.Dir {
					err := m.api.Cd(i.Name)
					if err != nil {
						m.setStatus("CD failed: "+err.Error(), true)
					} else {
						m.refreshFiles()
						m.setStatus("Changed directory", false)
					}
				} else {
					m.setStatus("Downloading...", false)
					err := m.api.Get(i.Name, i.Name)
					if err != nil {
						m.setStatus("Download failed: "+err.Error(), true)
					} else {
						m.setStatus("Downloaded "+i.Name+" successfully!", false)
					}
				}
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
	}

	var cmd tea.Cmd
	if !m.inputMode {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m *model) setStatus(msg string, isError bool) {
	m.statusMessage = msg
	m.statusIsError = isError
}

func (m *model) View() string {
	var s string

	if m.inputMode {
		s = fmt.Sprintf(
			"Upload File to Server\n\n%s\n\n(press Enter to submit, Esc to cancel)\n",
			m.textInput.View(),
		)
		s = docStyle.Render(s)
	} else {
		s = docStyle.Render(m.list.View())
	}

	if m.statusMessage != "" {
		style := successStyle
		if m.statusIsError {
			style = errorStyle
		}
		s += "\n" + style.Render(m.statusMessage)
	} else if !m.inputMode {
		s += "\n" + statusStyle.Render("Commands: [u]pload [r]efresh [q]uit [enter] open/download [/] filter")
	}

	return s
}

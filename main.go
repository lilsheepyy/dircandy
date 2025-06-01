package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	selectAction state = iota
	selectFiles
	selectDestination
	confirmRemove
	execute
	showResult
)

type actionItem struct {
	name string
}

func (a actionItem) Title() string       { return a.name }
func (a actionItem) Description() string { return "" }
func (a actionItem) FilterValue() string { return a.name }

type model struct {
	state           state
	action          string
	currentPath     string
	files           []os.DirEntry
	cursor          int
	selected        map[string]bool
	destinationPath string
	result          string
	err             error
	actionList      list.Model
}

func initialModel() model {
	items := []list.Item{
		actionItem{name: "Copy"},
		actionItem{name: "Move"},
		actionItem{name: "Remove"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 20, 7)
	l.Title = "Choose Action"

	return model{
		state:       selectAction,
		currentPath: ".", // Start wherever the user runs the app
		selected:    make(map[string]bool),
		actionList:  l,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case selectAction:
		var cmd tea.Cmd
		m.actionList, cmd = m.actionList.Update(msg)
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				if item, ok := m.actionList.SelectedItem().(actionItem); ok {
					m.action = item.name
					return m.loadFiles(m.currentPath, selectFiles)
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
		return m, cmd

	case selectFiles, selectDestination:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.files)-1 {
					m.cursor++
				}
			case "right", "enter":
				if m.files[m.cursor].IsDir() {
					newPath := filepath.Join(m.currentPath, m.files[m.cursor].Name())
					return m.loadFiles(newPath, m.state)
				}
			case "left", "backspace":
				parent := filepath.Dir(m.currentPath)
				if parent != m.currentPath { // Avoid looping when at root
					return m.loadFiles(parent, m.state)
				}
			case " ":
				if m.state == selectFiles {
					fullPath := filepath.Join(m.currentPath, m.files[m.cursor].Name())
					m.selected[fullPath] = !m.selected[fullPath]
				}
			case "tab":
				if m.state == selectFiles {
					if m.action == "Remove" {
						m.state = confirmRemove
					} else {
						return m.loadFiles(".", selectDestination)
					}
				} else if m.state == selectDestination {
					m.destinationPath = m.currentPath
					m.state = execute
					return m.runCommand()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

	case confirmRemove:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				m.state = execute
				return m.runCommand()
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

	case showResult:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) loadFiles(path string, nextState state) (tea.Model, tea.Cmd) {
	files, err := os.ReadDir(path)
	if err != nil {
		m.err = err
		m.state = showResult
		return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	m.files = files
	m.currentPath, _ = filepath.Abs(path) // Show absolute path for clarity
	m.cursor = 0
	m.state = nextState
	return m, nil
}

func (m model) runCommand() (tea.Model, tea.Cmd) {
	var cmd *exec.Cmd
	var paths []string
	for p, selected := range m.selected {
		if selected {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 {
		m.result = "No files selected."
		m.state = showResult
		return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}

	switch m.action {
	case "Copy":
		args := append([]string{"-r"}, paths...)
		args = append(args, m.destinationPath)
		cmd = exec.Command("cp", args...)
	case "Move":
		args := append(paths, m.destinationPath)
		cmd = exec.Command("mv", args...)
	case "Remove":
		args := append([]string{"-rf"}, paths...)
		cmd = exec.Command("rm", args...)
	}

	output, err := cmd.CombinedOutput()
	m.result = string(output)
	m.err = err
	m.state = showResult
	return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tea.Quit()
	})
}

func (m model) View() string {
	switch m.state {
	case selectAction:
		return m.actionList.View()
	case selectFiles:
		return m.renderFileList("Choose files (space to select, →/enter to open, ←/backspace to go up, tab to proceed)")
	case selectDestination:
		return m.renderFileList("Choose destination (→/enter to open, ←/backspace to go up, tab to confirm)")
	case confirmRemove:
		return "\nPress enter to confirm removal, or q to cancel."
	case showResult:
		status := "Success!"
		if m.err != nil {
			status = "Error: " + m.err.Error()
		}
		return fmt.Sprintf("\n== Result ==\n%s\n\n%s\n\nExiting in 2 seconds...", status, m.result)
	default:
		return "Loading..."
	}
}

func (m model) renderFileList(title string) string {
	s := fmt.Sprintf("\nCurrent Path: %s\n%s\n", m.currentPath, title)
	for i, file := range m.files {
		cursor := "  "
		if m.cursor == i {
			cursor = "👉"
		}
		fullPath := filepath.Join(m.currentPath, file.Name())
		selected := "  "
		if m.state == selectFiles && m.selected[fullPath] {
			selected = "✅"
		}
		tag := ""
		if file.IsDir() {
			tag = " [DIR]"
		}
		info, _ := file.Info()
		size := ""
		if !file.IsDir() {
			size = fmt.Sprintf(" (%d bytes)", info.Size())
		}
		s += fmt.Sprintf("%s%s %s%s%s\n", cursor, selected, file.Name(), tag, size)
	}
	s += "\nControls: ↑/↓ to navigate, space to select files, →/enter to open, ←/backspace to go up, tab to proceed, q to quit"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

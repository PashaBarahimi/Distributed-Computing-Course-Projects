package mainmenu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

type Model struct {
	choices       []string
	cursor        int
	Choice        string
	statusMessage string
}

func New(choices []string, statusMessage string) Model {
	return Model{choices: choices, statusMessage: statusMessage}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			m.Choice = m.choices[m.cursor]
			return m, tea.Quit

		case tea.KeyDown:
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case tea.KeyUp:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}

		default:
			break
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(m.statusMessage + "\n")
	s.WriteString("Choose the desired option:\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString(checkboxStyle.Render("[x] " + m.choices[i]))
		} else {
			s.WriteString("[ ] " + m.choices[i])
		}
		s.WriteString("\n")
	}
	s.WriteString("\n(press esc to quit)\n")

	return s.String()
}

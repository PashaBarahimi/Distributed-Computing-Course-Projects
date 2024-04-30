package eventcreator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
}

func (m Model) GetName() string {
	return m.inputs[0].Value()
}

func (m Model) GetTime() time.Time {
	str := m.inputs[1].Value()
	if str == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.DateTime, str)
	return t
}

func (m Model) GetTotalTickets() int {
	str := m.inputs[2].Value()
	if str == "" {
		return 0
	}
	t, _ := strconv.Atoi(str)
	return t
}

func validateTime(s string) error {
	if len(s) > 19 {
		return fmt.Errorf("too long")
	}

	f := "dddd-dd-dd dd:dd:dd"
	for i := 0; i < len(s); i++ {
		if f[i] == '-' || f[i] == ':' || f[i] == ' ' {
			if s[i] != f[i] {
				return fmt.Errorf("invalid format")
			}
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return fmt.Errorf("invalid character")
		}
	}

	if len(s) >= 4 {
		year, _ := strconv.Atoi(s[:4])
		if year < 1970 {
			return fmt.Errorf("invalid year")
		}
	}

	if len(s) >= 7 {
		month, _ := strconv.Atoi(s[5:7])
		if month < 1 || month > 12 {
			return fmt.Errorf("invalid month")
		}
	}

	if len(s) >= 10 {
		_, err := time.Parse(time.DateOnly, s[:10])
		if err != nil {
			return fmt.Errorf("invalid date")
		}
	}

	if len(s) >= 13 {
		hour, _ := strconv.Atoi(s[11:13])
		if hour < 0 || hour > 23 {
			return fmt.Errorf("invalid hour")
		}
	}

	if len(s) >= 16 {
		minute, _ := strconv.Atoi(s[14:16])
		if minute < 0 || minute > 59 {
			return fmt.Errorf("invalid minute")
		}
	}

	if len(s) == 19 {
		_, err := time.Parse(time.DateTime, s)
		if err != nil {
			return fmt.Errorf("invalid time")
		}
	}

	return nil
}

func validateTickets(s string) error {
	if s == "" {
		return nil
	}

	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid number")
	}

	return nil
}

func New() Model {
	m := Model{
		inputs: make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Prompt = "> Name: "
			t.Placeholder = "Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Prompt = "> Time: "
			t.Placeholder = time.DateTime
			t.CharLimit = 64
			t.Validate = validateTime
		case 2:
			t.Prompt = "> Total Tickets: "
			t.Placeholder = "Total Tickets"
			t.CharLimit = 10
			t.Validate = validateTickets
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			t := msg.Type

			if t == tea.KeyEnter && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			if t == tea.KeyUp || t == tea.KeyShiftTab {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	_, _ = fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

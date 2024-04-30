package loadspinner

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error
type doneMsg struct{}

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

func waitCmd(ch chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		return doneMsg{}
	}
}

type Model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	ch       chan struct{}
	message  string
}

func New(ch chan struct{}, message string) Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = spinnerStyle
	return Model{
		spinner: s,
		ch:      ch,
		err:     nil,
		message: message,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitCmd(m.ch))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		default:
			return m, nil
		}

	case doneMsg:
		m.quitting = true
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.message)
	if m.quitting {
		return str + "\n"
	}
	return str
}

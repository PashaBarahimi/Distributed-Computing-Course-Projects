package ticketselector

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

func validateNumber(input string, maxNum int) error {
	num, err := strconv.Atoi(input)
	if err != nil {
		return errMsg(fmt.Errorf("invalid number"))
	}
	if num > maxNum || num <= 0 {
		return errMsg(fmt.Errorf("number is out of range"))
	}
	return nil
}

type Model struct {
	textInput     textinput.Model
	err           error
	ChosenTickets int
}

func New(availableTickets int) Model {
	ti := textinput.New()
	ti.Placeholder = strconv.Itoa(availableTickets)
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.Validate = func(input string) error {
		return validateNumber(input, availableTickets)
	}

	return Model{
		textInput: ti,
		err:       nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.textInput.Value() == "" {
				return m, nil
			}
			num, _ := strconv.Atoi(m.textInput.Value())
			m.ChosenTickets = num
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf(
		"Enter the number of tickets:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

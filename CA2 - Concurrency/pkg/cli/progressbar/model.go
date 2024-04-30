package progressbar

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxWidth                   = 80
	percentageIncreaseValue    = 0.25
	percentageIncreaseDuration = time.Millisecond * 500
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(percentageIncreaseDuration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type Model struct {
	Progress      progress.Model
	paddingWidth  int
	paddingHeight int
}

func New() Model {
	return Model{
		Progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.paddingWidth = (msg.Width - maxWidth) / 2
		if m.paddingWidth < 0 {
			m.paddingWidth = 0
		}
		m.paddingHeight = (msg.Height - 1) / 2
		if m.paddingHeight < 0 {
			m.paddingHeight = 0
		}

		m.Progress.Width = msg.Width - m.paddingWidth*2 - 4
		if m.Progress.Width > maxWidth {
			m.Progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if m.Progress.Percent() == 1.0 {
			return m, tea.Quit
		}

		cmd := m.Progress.IncrPercent(percentageIncreaseValue)
		return m, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m Model) View() string {
	padWidth := strings.Repeat(" ", m.paddingWidth)
	padHeight := strings.Repeat("\n", m.paddingHeight)
	return padHeight + padWidth + m.Progress.View() + "\n"
}

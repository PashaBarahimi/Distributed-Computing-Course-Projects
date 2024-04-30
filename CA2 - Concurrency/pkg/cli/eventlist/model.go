package eventlist

import (
	"dist-concurrency/pkg/event"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
)

type Model struct {
	ChosenItem string
	list       list.Model
	keys       *listKeyMap
}

func New(events []event.Event) Model {
	listKeys := newListKeyMap()
	items := make([]list.Item, 0, len(events))
	for _, e := range events {
		items = append(items, Item{
			Name:             e.Name,
			Id:               e.ID,
			AvailableTickets: e.AvailableTickets,
			TotalTickets:     e.TotalTickets,
			Date:             e.Date,
		})
	}

	itemsList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	itemsList.Title = "Events"
	itemsList.Styles.Title = titleStyle
	itemsList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.chooseItem,
		}
	}
	itemsList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.chooseItem,
		}
	}

	return Model{
		ChosenItem: "",
		list:       itemsList,
		keys:       listKeys,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.chooseItem):
			chosen := m.list.SelectedItem().(Item)
			m.ChosenItem = chosen.Id
			return m, tea.Quit
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return appStyle.Render(m.list.View())
}

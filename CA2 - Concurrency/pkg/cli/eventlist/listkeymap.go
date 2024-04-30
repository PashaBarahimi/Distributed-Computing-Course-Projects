package eventlist

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	chooseItem key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		chooseItem: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose Item"),
		),
	}
}

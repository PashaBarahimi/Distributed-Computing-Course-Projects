package eventlist

import (
	"strconv"
	"time"
)

type Item struct {
	Name             string
	Id               string
	AvailableTickets int
	TotalTickets     int
	Date             time.Time
}

func (i Item) Title() string {
	return i.Name + " - " + i.Date.Format("2006-01-02 15:04")
}

func (i Item) Description() string {
	return "Available tickets: " + strconv.Itoa(i.AvailableTickets) + "/" + strconv.Itoa(i.TotalTickets)
}

func (i Item) FilterValue() string {
	return i.Name
}

package event

import "time"

type Event struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Date             time.Time `json:"date"`
	TotalTickets     int       `json:"totalTickets"`
	AvailableTickets int       `json:"availableTickets"`
}

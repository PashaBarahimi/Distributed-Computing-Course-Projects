package ticketservice

import (
	"fmt"
	"sync"
	"time"

	"dist-concurrency/pkg/event"
	
	"github.com/google/uuid"
)

type TicketService struct {
	events sync.Map
}

func (ts *TicketService) CreateEvent(name string, date time.Time, totalTickets int) (*event.Event, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	e := &event.Event{
		ID:               id.String(),
		Name:             name,
		Date:             date,
		TotalTickets:     totalTickets,
		AvailableTickets: totalTickets,
	}
	ts.events.Store(e.ID, e)
	return e, nil
}

func (ts *TicketService) ListEvents() []*event.Event {
	var events []*event.Event
	ts.events.Range(func(key, value interface{}) bool {
		e := value.(*event.Event)
		events = append(events, e)
		return true
	})
	return events
}

func (ts *TicketService) BookTickets(eventID string, numTickets int) ([]string, error) {
	// TODO: implement concurrency control

	e, ok := ts.events.Load(eventID)
	if !ok {
		return nil, fmt.Errorf("event not found")
	}

	ev := e.(*event.Event)
	if ev.AvailableTickets < numTickets {
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		ticketID := id.String()
		ticketIDs = append(ticketIDs, ticketID)
		// TODO: store the ticket in a separate data structure if needed
	}

	ev.AvailableTickets -= numTickets
	ts.events.Store(eventID, ev) // FIXME: do we need to do this?

	return ticketIDs, nil
}

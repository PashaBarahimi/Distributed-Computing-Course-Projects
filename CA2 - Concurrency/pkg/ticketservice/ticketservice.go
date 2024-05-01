package ticketservice

import (
	"fmt"
	"sync"
	"time"

	"dist-concurrency/pkg/event"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type TicketService struct {
	events  sync.Map
	tickets sync.Map
	mu      sync.RWMutex
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
	ts.mu.Lock()
	ts.events.Store(e.ID, e)
	ts.mu.Unlock()
	return e, nil
}

func (ts *TicketService) ListEvents() []*event.Event {
	var events []*event.Event
	ts.mu.RLock()

	ts.events.Range(func(key, value interface{}) bool {
		e, ok := value.(*event.Event)
		if !ok {
			log.Errorf("invalid event: %v", value)
		}
		events = append(events, e)
		return true
	})

	ts.mu.RUnlock()
	log.Infof("Listing %d events", len(events))
	return events
}

func (ts *TicketService) BookTickets(eventID string, numTickets int) ([]string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	e, ok := ts.events.Load(eventID)
	if !ok {
		return nil, fmt.Errorf("event not found")
	}

	ev, ok := e.(*event.Event)
	if !ok {
		return nil, fmt.Errorf("invalid event")
	}
	if ev.AvailableTickets < numTickets {
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		id := uuid.New()
		ticketID := id.String()
		ticketIDs = append(ticketIDs, ticketID)
		// TODO: store the ticket in a separate data structure if needed
	}

	ev.AvailableTickets -= numTickets
	log.Infof("Booked %d tickets for event %s", numTickets, ev.Name)

	for _, ticketID := range ticketIDs {
		log.Infof("Storing ticket %s for event %s", ticketID, ev.Name)
		ts.tickets.Store(ticketID, eventID)
	}

	return ticketIDs, nil
}

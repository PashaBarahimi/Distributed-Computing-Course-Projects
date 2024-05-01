package ticketservice

import (
	"fmt"
	"sync"
	"time"

	"dist-concurrency/pkg/cache"
	"dist-concurrency/pkg/event"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type TicketService struct {
	events  sync.Map
	tickets sync.Map
	mu      sync.RWMutex
	cache   *cache.Cache
}

func New() *TicketService {
	ts := &TicketService{}
	ts.cache = cache.New(&ts.events, &ts.mu, 10)
	return ts
}

func (ts *TicketService) CreateEvent(name string, date time.Time, totalTickets int) (*event.Event, error) {
	id := uuid.New()
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
	e := ts.cache.GetEvent(eventID)
	if e == nil {
		return nil, fmt.Errorf("event not found")
	}

	e.Mu.Lock()
	defer e.Mu.Unlock()
	ev := e.Event

	if ev.AvailableTickets < numTickets {
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		id := uuid.New()
		ticketID := id.String()
		ticketIDs = append(ticketIDs, ticketID)
	}

	ev.AvailableTickets -= numTickets
	log.Infof("Booked %d tickets for event %s", numTickets, ev.Name)

	for _, ticketID := range ticketIDs {
		log.Infof("Storing ticket %s for event %s", ticketID, ev.Name)
		ts.tickets.Store(ticketID, eventID)
	}

	return ticketIDs, nil
}

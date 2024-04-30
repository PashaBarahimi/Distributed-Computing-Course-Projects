package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"dist-concurrency/pkg/cli/eventlist"
	"dist-concurrency/pkg/cli/loadspinner"
	"dist-concurrency/pkg/cli/logport"
	"dist-concurrency/pkg/cli/mainmenu"
	"dist-concurrency/pkg/cli/progressbar"
	"dist-concurrency/pkg/cli/ticketselector"
	"dist-concurrency/pkg/event"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
)

const (
	events string = "Events"
	logs   string = "Logs"
	quit   string = "Quit"

	defaultHost = "localhost"
	defaultPort = 8080

	listEventsPath     = "/events"
	reserveTicketsPath = "/reserve"
)

var (
	mainMenuChoices = []string{events, logs, quit}

	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()

	logBuffer = strings.Builder{}

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	port int
	host string
)

func init() {
	log.SetPrefix("Client")
	log.SetTimeFormat(time.TimeOnly)
	log.SetLevel(log.DebugLevel)
	log.SetOutput(&logBuffer)
}

func getBaseUrl() string {
	return fmt.Sprintf("http://%s:%d", host, port)
}

func sendHttpRequest(req *http.Request) (resp *http.Response, err error) {
	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code %s", resp.Status)
	}
	return resp, nil
}

func loadProgressBar() {
	log.Info("Loading progress bar")
	pbModel := progressbar.New()
	_, err := tea.NewProgram(pbModel, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error loading progress bar: %v", err)
		return
	}
}

func loadMainMenu() {
	status := ""
	for {
		log.Info("Loading main menu")
		menuModel := mainmenu.New(mainMenuChoices, status)
		m, err := tea.NewProgram(menuModel, tea.WithAltScreen()).Run()
		if err != nil {
			log.Errorf("Error loading main menu: %v", err)
			return
		}

		menuModel, _ = m.(mainmenu.Model)
		switch menuModel.Choice {
		case events:
			status = loadEvents()
		case logs:
			loadLogs()
			status = ""
		default:
			return
		}
	}
}

func loadEvents() (status string) {
	ch := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go loadSpinner(ch, &wg, "Retrieving events...")
	events, err := getEvents()
	ch <- struct{}{}
	wg.Wait()
	if err != nil {
		log.Errorf("Error retrieving events: %v", err)
		status = red(err.Error())
		return
	}
	log.Infof("Loading events list with %d events", len(events))
	eventsModel := eventlist.New(events)
	m, err := tea.NewProgram(eventsModel, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error loading events list: %v", err)
		return
	}

	eventsModel, _ = m.(eventlist.Model)
	id := eventsModel.ChosenItem
	if id != "" {
		e := findEventByID(events, id)
		tickets := promptForTickets(e)
		if tickets > 0 {
			wg.Add(1)
			go loadSpinner(ch, &wg, "Reserving tickets...")
			err = reserveTickets(e, tickets)
			if err != nil {
				log.Errorf("Error reserving tickets: %v", err)
				status = red("Failed to reserve tickets, check logs")
			} else {
				log.Info("Tickets reserved successfully")
				status = green("Tickets reserved successfully")
			}
			ch <- struct{}{}
			wg.Wait()
		} else {
			log.Warn("No tickets selected")
			status = yellow("No tickets selected")
		}
	} else {
		log.Warn("No event selected")
		status = yellow("No event selected")
	}
	return
}

func loadSpinner(ch chan struct{}, wg *sync.WaitGroup, message string) {
	log.Infof("Loading spinner with message: %s", message)
	spinnerModel := loadspinner.New(ch, message)
	_, err := tea.NewProgram(spinnerModel, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error loading spinner: %v", err)
	}
	wg.Done()
}

func promptForTickets(e event.Event) int {
	log.Infof("Prompting user for tickets for event %s", e.Name)
	ticketsModel := ticketselector.New(e.AvailableTickets)
	m, err := tea.NewProgram(ticketsModel, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error prompting user for tickets: %v", err)
		return 0
	}

	ticketsModel, _ = m.(ticketselector.Model)
	log.Infof("User chose %d tickets", ticketsModel.ChosenTickets)
	return ticketsModel.ChosenTickets
}

func loadLogs() {
	log.Info("Loading logs...")
	logs := getLogs()
	logsModel := logport.New(logs)
	_, err := tea.NewProgram(logsModel, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		log.Errorf("Error loading logs: %v", err)
	}
}

func getEvents() ([]event.Event, error) {
	log.Info("Retrieving events...")
	req, err := http.NewRequest("GET", getBaseUrl()+listEventsPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := sendHttpRequest(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Error closing response body: %v", err)
		}
	}(resp.Body)

	var events []event.Event
	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		return nil, err
	}

	log.Infof("Retrieved %d events", len(events))
	return events, nil
}

func findEventByID(events []event.Event, id string) event.Event {
	for _, e := range events {
		if e.ID == id {
			return e
		}
	}
	return event.Event{}
}

func reserveTickets(e event.Event, tickets int) error {
	log.Infof("Reserving %d tickets for event %s", tickets, e.Name)
	req, err := http.NewRequest("POST", getBaseUrl()+reserveTicketsPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("eventId", e.ID)
	q.Add("tickets", fmt.Sprintf("%d", tickets))
	req.URL.RawQuery = q.Encode()
	resp, err := sendHttpRequest(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Error closing response body: %v", err)
		}
	}(resp.Body)
	var ticketIds []string
	err = json.NewDecoder(resp.Body).Decode(&ticketIds)
	if err != nil {
		return err
	}
	log.Infof("Reserved tickets with IDs: %v", ticketIds)
	return nil
}

func getLogs() string {
	s := logBuffer.String()
	s = strings.ReplaceAll(s, "DEBU", blue("DEBU"))
	s = strings.ReplaceAll(s, "INFO", green("INFO"))
	s = strings.ReplaceAll(s, "WARN", yellow("WARN"))
	s = strings.ReplaceAll(s, "ERRO", red("ERRO"))
	return s
}

func main() {
	log.Info("Starting client...")

	portPtr := flag.Int("port", defaultPort, "Server port number")
	hostPtr := flag.String("host", defaultHost, "Server host address")
	flag.Parse()
	port = *portPtr
	host = *hostPtr
	log.Infof("Connecting to server at %s:%d", host, port)

	loadProgressBar()
	loadMainMenu()
}

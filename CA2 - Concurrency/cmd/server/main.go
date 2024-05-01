package main

import (
	"dist-concurrency/pkg/cli/eventcreator"
	"dist-concurrency/pkg/cli/eventlist"
	"dist-concurrency/pkg/cli/logport"
	"dist-concurrency/pkg/cli/mainmenu"
	"dist-concurrency/pkg/cli/progressbar"
	"dist-concurrency/pkg/event"
	"dist-concurrency/pkg/ticketservice"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"

	"github.com/charmbracelet/log"
)

const (
	events   string = "Events"
	addEvent string = "Add Event"
	logs     string = "Logs"
	quit     string = "Quit"

	defaultHost = "127.0.0.1"
	defaultPort = 8080

	listEventsPath     = "/events"
	reserveTicketsPath = "/reserve"
)

var (
	mainMenuChoices = []string{events, addEvent, logs, quit}

	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()

	logBuffer = strings.Builder{}

	service = ticketservice.TicketService{}

	host string
	port int
)

func init() {
	log.SetPrefix("Server")
	log.SetTimeFormat(time.TimeOnly)
	log.SetLevel(log.DebugLevel)
	log.SetOutput(&logBuffer)
}

func writeBody(w http.ResponseWriter, body []byte) {
	write, err := w.Write(body)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Debugf("Response: %v", http.StatusInternalServerError)
	} else {
		log.Debugf("Wrote %d bytes", write)
	}
}

func listEvents(w http.ResponseWriter, r *http.Request) {
	log.Info("Listing events...")
	events := service.ListEvents()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(events)
	writeBody(w, body)
	log.Debugf("Response: %v", http.StatusOK)
	time.Sleep(time.Second)
}

func reserveTickets(w http.ResponseWriter, r *http.Request) {
	log.Info("Reserving tickets...")
	eventID := r.URL.Query().Get("eventId")
	numTickets, err := strconv.Atoi(r.URL.Query().Get("tickets"))
	if err != nil {
		log.Errorf("Error parsing tickets: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		body := []byte(err.Error())
		writeBody(w, body)
		log.Debugf("Response: %v", http.StatusBadRequest)
		return
	}

	ticketIDs, err := service.BookTickets(eventID, numTickets)
	if err != nil {
		log.Errorf("Error booking tickets: %v", err)
		w.WriteHeader(http.StatusTeapot)
		body := []byte(err.Error())
		writeBody(w, body)
		log.Debugf("Response: %v", http.StatusTeapot)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(ticketIDs)
	writeBody(w, body)
	log.Debugf("Response: %v", http.StatusOK)
	time.Sleep(time.Second)
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
	for {
		log.Info("Loading main menu")
		menuModel := mainmenu.New(mainMenuChoices, "")
		m, err := tea.NewProgram(menuModel, tea.WithAltScreen()).Run()
		if err != nil {
			log.Errorf("Error loading main menu: %v", err)
			return
		}

		menuModel, _ = m.(mainmenu.Model)
		switch menuModel.Choice {
		case events:
			loadEvents()
		case addEvent:
			loadAddEvent()
		case logs:
			loadLogs()
		default:
			os.Exit(0)
		}
	}
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

func loadEvents() {
	log.Info("Loading events...")
	var events []event.Event
	for _, e := range service.ListEvents() {
		events = append(events, *e)
	}
	eventsModel := eventlist.New(events)
	_, err := tea.NewProgram(eventsModel, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error loading events: %v", err)
	}
}

func loadAddEvent() {
	log.Info("Loading add event...")
	model := eventcreator.New()
	m, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		log.Errorf("Error loading add event: %v", err)
		return
	}

	model, _ = m.(eventcreator.Model)
	if model.GetName() == "" || model.GetTime().Year() == 0 || model.GetTotalTickets() == 0 {
		log.Warn("Invalid event")
		return
	}

	createEvent, err := service.CreateEvent(model.GetName(), model.GetTime(), model.GetTotalTickets())
	if err != nil {
		log.Errorf("Error creating event: %v", err)
		return
	}
	log.Infof("Created event: %v", createEvent)
}

func getLogs() string {
	s := logBuffer.String()
	s = strings.ReplaceAll(s, "DEBU", blue("DEBU"))
	s = strings.ReplaceAll(s, "INFO", green("INFO"))
	s = strings.ReplaceAll(s, "WARN", yellow("WARN"))
	s = strings.ReplaceAll(s, "ERRO", red("ERRO"))
	return s
}

func handleCli() {
	loadProgressBar()
	loadMainMenu()
}

func main() {
	log.Info("Starting server...")

	portPtr := flag.Int("port", defaultPort, "Server port number")
	hostPtr := flag.String("host", defaultHost, "Server host address")
	flag.Parse()
	port = *portPtr
	host = *hostPtr
	log.Infof("Listening on %s:%d", host, port)

	http.HandleFunc(listEventsPath, listEvents)
	http.HandleFunc(reserveTicketsPath, reserveTickets)

	go handleCli()

	err := http.ListenAndServe(host+":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

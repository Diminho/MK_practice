package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"sync"

	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/websocket"
)

type EventPlaces struct {
	Event          string   `json:"event"`
	BookedPlaces   []string `json:"places"`
	LastActedPlace string   `json:"lastActedPlace"`
	Action         string   `json:"action"`
	UserAddr       string   `json:"userAddr"`
	ErrorCode      int      `json:"errorCode"`
}

type EventSystemMessage struct {
	Message        string `json:"sysMessage"`
	MessageType    int    `json:"messageType"`
	Event          string `json:"event"`
	LastActedPlace string `json:"lastActedPlace"`
}

type EventPlacesRow struct {
	PlaceIdentity string
	IsBooked      int
	IsBought      int
	UserID        string
}

type EventPlacesTemplate struct {
	EventPlacesRows []EventPlacesRow
	EventID         int
}

var eventClients = make(map[string][]*websocket.Conn)
var broadcast = make(chan EventPlaces)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//not the best practice to use global, but left for simplicity
var db *sql.DB

var mu = &sync.Mutex{}

func main() {
	var dbErr error
	db, dbErr = sql.Open("mysql", "root:@/ticket_booking")
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	defer db.Close()

	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/event", handleEvent)
	go handlePlaceBookings()

	log.Println("Server started. Port: 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	populateTemplate(queryForAllPlacesInEvent(), w, "../public/event.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}

	_, message, err := wsConnection.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	event := string(message)

	clients, ok := eventClients[event]

	if !ok {
		client := []*websocket.Conn{wsConnection}
		eventClients[event] = client
	} else {
		clients = append(clients, wsConnection)
		eventClients[event] = clients
	}

	// Defering here since we have defined eventClients for this particular connection
	defer func() {
		deleteClient(eventClients[event], indexOfClient(wsConnection, eventClients[event]))
		wsConnection.Close()
	}()

	for {
		var places EventPlaces
		//imagine data is not corrupted or mixed up
		erro := wsConnection.ReadJSON(&places)
		if erro != nil {
			log.Printf("error: %v", erro)
			return
		}

		mu.Lock()
		code := processPlace(places.LastActedPlace, places.Action, wsConnection.RemoteAddr().String())
		mu.Unlock()
		places.ErrorCode = code
		occupied := queryForOccupiedPlacesInEvent()

		for _, occupiedPlace := range occupied {
			if !inStringSlice(places.BookedPlaces, occupiedPlace.PlaceIdentity) {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
			}
		}

		places.UserAddr = wsConnection.RemoteAddr().String()

		broadcast <- places
	}
}

func handlePlaceBookings() {
	for places := range broadcast {
		for _, client := range eventClients[places.Event] {
			if places.UserAddr == client.RemoteAddr().String() {
				er := client.WriteJSON(getEventSystemMessage(places.ErrorCode, places.Event, places.LastActedPlace))

				if er != nil {
					log.Printf("error: %v", er)
				}
				continue
			}
			er := client.WriteJSON(places)

			if er != nil {
				log.Printf("error: %v", er)
			}
		}
	}
}

func populateTemplate(data interface{}, w http.ResponseWriter, tmplFile string) {
	tmpl, _ := template.ParseFiles(tmplFile)
	tmpl.Execute(w, data)
}

func inStringSlice(input []string, needle string) bool {
	for _, elem := range input {
		if elem == needle {
			return true
		}
	}

	return false
}

func deleteClient(a []*websocket.Conn, index int) {
	copy(a[index:], a[index+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
}

func indexOfClient(conn *websocket.Conn, data []*websocket.Conn) int {
	for k, value := range data {
		if conn == value {
			return k
		}
	}
	return -1
}

func getEventSystemMessage(code int, event string, place string) EventSystemMessage {
	var message string

	switch code {
	case 0:
		message = "Success"
	case 1:
		message = fmt.Sprintf("Sorry, this [%s] have been already booked", place)
	}

	return EventSystemMessage{Message: message, MessageType: code, Event: event, LastActedPlace: place}
}

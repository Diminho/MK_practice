package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Diminho/MK_practice/config"
	"github.com/Diminho/MK_practice/models"
	"github.com/gorilla/websocket"
)

type App struct {
	db           models.Database
	broadcast    chan models.EventPlaces
	eventClients map[string][]*websocket.Conn
}

type EventSystemMessage struct {
	Message        string `json:"sysMessage"`
	MessageType    int    `json:"messageType"`
	Event          string `json:"event"`
	LastActedPlace string `json:"lastActedPlace"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//not the best practice to use global, but left for simplicity

var mu = &sync.Mutex{}

func main() {
	db, dbErr := models.Connect(config.User, config.Host, config.DbName)
	if dbErr != nil {
		log.Panic(dbErr)
	}

	app := &App{
		db:           db,
		eventClients: make(map[string][]*websocket.Conn),
		broadcast:    make(chan models.EventPlaces),
	}

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", app.handleConnections)
	http.HandleFunc("/event", app.handleEvent)
	go app.handlePlaceBookings()

	log.Println("Server started. Port: 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func (app *App) handleEvent(w http.ResponseWriter, r *http.Request) {
	populateTemplate(app.db.QueryForAllPlacesInEvent(), w, "public/event.html")
}

func (app *App) handleConnections(w http.ResponseWriter, r *http.Request) {
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

	clients, ok := app.eventClients[event]

	if !ok {
		client := []*websocket.Conn{wsConnection}
		app.eventClients[event] = client
	} else {
		clients = append(clients, wsConnection)
		app.eventClients[event] = clients
	}

	// Defering here since we have defined eventClients for this particular connection
	defer func() {
		deleteClient(app.eventClients[event], indexOfClient(wsConnection, app.eventClients[event]))
		wsConnection.Close()
	}()

	for {
		var places models.EventPlaces
		//imagine data is not corrupted or mixed up
		erro := wsConnection.ReadJSON(&places)
		if erro != nil {
			log.Printf("error: %v", erro)
			return
		}

		mu.Lock()
		code := app.db.ProcessPlace(places.LastActedPlace, places.Action, wsConnection.RemoteAddr().String())
		mu.Unlock()
		places.ErrorCode = code
		occupied := app.db.QueryForOccupiedPlacesInEvent()

		for _, occupiedPlace := range occupied {
			if !inStringSlice(places.BookedPlaces, occupiedPlace.PlaceIdentity) {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
			}
		}

		places.UserAddr = wsConnection.RemoteAddr().String()

		app.broadcast <- places
	}
}

func (app *App) handlePlaceBookings() {
	for places := range app.broadcast {
		for _, client := range app.eventClients[places.Event] {
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

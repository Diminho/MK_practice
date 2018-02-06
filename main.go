package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	logjson "github.com/Diminho/MK_practice/simplelog/handlers/json"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

type App struct {
	broadcast      chan models.EventPlaces
	eventClients   map[string][]*websocket.Conn
	facebookState  string
	facebookConfig *oauth2.Config
	slog           *simplelog.Log
	db             func() *models.DB
}

type EventSystemMessage struct {
	Message        string `json:"sysMessage"`
	MessageType    int    `json:"messageType"`
	Event          string `json:"event"`
	LastActedPlace string `json:"lastActedPlace"`
	BookTime       int    `json:"bookTime"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mu = &sync.Mutex{}

func main() {

	file, fileErr := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if fileErr != nil {
		log.Fatal("Failed to open log file: ", fileErr)
	}
	defer file.Close()

	slog := simplelog.NewLog(file)
	slog.SetHandler(logjson.New(slog))

	// db, dbErr := models.Connect()
	// if dbErr != nil {
	// 	slog.Fatal(dbErr)
	// }

	// app := &App{
	// 	eventClients: make(map[string][]*websocket.Conn),
	// 	broadcast:    make(chan models.EventPlaces),
	// 	facebookConfig: &oauth2.Config{
	// 		ClientID:     config.ClientID,
	// 		ClientSecret: config.ClientSecret,
	// 		RedirectURL:  config.RedirectURL,
	// 		Scopes:       config.Scopes,
	// 		Endpoint:     config.Endpoint,
	// 	},
	// 	facebookState: "MK_PRACTICE",
	// 	db:            db.DBInstance(),
	// 	slog:          slog,
	// }

	// fs := http.FileServer(http.Dir("./public"))
	// http.Handle("/", fs)

	// http.HandleFunc("/ws", app.handleConnections)
	// http.HandleFunc("/event", app.handleEvent)
	// http.HandleFunc("/facebook_login", app.handleFacebookLogin)
	// http.HandleFunc("/facebookCallback", app.handleFacebookCallback)
	// go app.handlePlaceBookings()

	// //simple graceful shutdown
	// var gracefulShut = make(chan os.Signal)
	// signal.Notify(gracefulShut, syscall.SIGTERM)
	// signal.Notify(gracefulShut, syscall.SIGINT)
	// go func() {
	// 	sig := <-gracefulShut
	// 	slog.Infof("caught signal: %+v", sig)
	// 	slog.Info("Wait for 5 second to finish processing")
	// 	time.Sleep(5 * time.Second)
	// 	os.Exit(0)
	// }()

	// slog.Info("Server running on port :8000")
	// err := http.ListenAndServe(":8000", nil)
	// if err != nil {
	// 	slog.Fatal(err)
	// }

}

func (app *App) handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != app.facebookState {
		app.slog.Infof("invalid oauth state, expected '%s', got '%s'\n", app.facebookState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := app.facebookConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		app.slog.Infof("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?fields=email,name,id&access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	var user models.User
	json.Unmarshal(response, &user)
	if !app.db().UserExists(user.Email) {
		app.db().AddNewUser(&user)

	}

	expiration := time.Now().Add(60 * time.Second)
	//Warning! used to set email in Cookie just for simplicity.
	cookie := http.Cookie{Name: "ticket_booking", Value: user.Email, Expires: expiration}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, r.Referer(), http.StatusTemporaryRedirect)
}

func (app *App) handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	authURL, err := url.Parse(app.facebookConfig.Endpoint.AuthURL)
	if err != nil {
		app.slog.Error(err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", app.facebookConfig.ClientID)
	parameters.Add("scope", strings.Join(app.facebookConfig.Scopes, " "))
	parameters.Add("redirect_uri", app.facebookConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", app.facebookState)
	authURL.RawQuery = parameters.Encode()
	url := authURL.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *App) handleEvent(w http.ResponseWriter, r *http.Request) {
	tmplData := app.db().AllPlacesInEvent()
	tmplData.Request = r

	cookie, ok := isLogged(r)
	fmt.Println(cookie)
	if !ok {
		tmplData.UserInfo["isLogged"] = "0"
	} else {
		user, _ := app.db().FindUserByEmail(cookie.Value)
		tmplData.UserInfo["isLogged"] = "1"
		tmplData.UserInfo["name"] = user.Name
		tmplData.UserInfo["id"] = user.ID
	}

	populateTemplate(tmplData, w, "public/event.html")
}

func (app *App) handleConnections(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		app.slog.Fatal(err)
	}

	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		app.slog.Error(err)
		return
	}

	event := string(msg)

	clients, ok := app.eventClients[event]

	if !ok {
		client := []*websocket.Conn{wsConn}
		app.eventClients[event] = client
	} else {
		clients = append(clients, wsConn)
		app.eventClients[event] = clients
	}

	// Defering here since we have defined eventClients for this particular connection
	defer func() {
		deleteClient(app.eventClients[event], indexOfClient(wsConn, app.eventClients[event]))
		wsConn.Close()
	}()

	for {
		var places models.EventPlaces
		//imagine data is not corrupted or mixed up
		err := wsConn.ReadJSON(&places)
		if err != nil {
			app.slog.Error(err)
			return
		}

		mu.Lock()
		code := app.db().ProcessPlace(&places, buildUserIdentity(wsConn.RemoteAddr().String()))
		mu.Unlock()
		places.ErrorCode = code
		occupied := app.db().OccupiedPlacesInEvent()

		for _, occupiedPlace := range occupied {
			if !inStringSlice(places.BookedPlaces, occupiedPlace.PlaceIdentity) {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
			}
		}

		places.UserAddr = wsConn.RemoteAddr().String()

		app.broadcast <- places
	}
}

func (app *App) handlePlaceBookings() {

	for places := range app.broadcast {
		for _, client := range app.eventClients[places.Event] {
			if client == nil {
				continue
			}
			if places.UserAddr == client.RemoteAddr().String() {
				err := client.WriteJSON(getEventSystemMessage(places.ErrorCode, places.Event, places.LastActedPlace))

				if err != nil {
					app.slog.Error(err)
				}
				continue
			}
			err := client.WriteJSON(places)

			if err != nil {
				app.slog.Error(err)
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

func deleteClient(a []*websocket.Conn, i int) {
	copy(a[i:], a[i+1:])
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

	return EventSystemMessage{Message: message, MessageType: code, Event: event, LastActedPlace: place, BookTime: models.BookTime}
}

func buildUserIdentity(remoteAddr string) string {
	host, port, _ := net.SplitHostPort(remoteAddr)
	userID := fmt.Sprintf("%s_%s", host, port)
	return userID
}

func isLogged(r *http.Request) (*http.Cookie, bool) {
	var isLogged bool
	cookie, errorCookie := r.Cookie("ticket_booking")
	if errorCookie == nil {
		isLogged = true
	}
	return cookie, isLogged
}

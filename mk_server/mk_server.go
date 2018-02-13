package mk_server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Diminho/MK_practice/app"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

//embedding App to enable declaring handlers
type WraperApp struct {
	*app.App
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(app *app.App) {

	wApp := &WraperApp{app}

	s := &http.Server{Addr: ":8000", Handler: LoadRoutes(wApp)}

	go wApp.handlePlaceBookings()

	//simple graceful shutdown
	var gracefulShut = make(chan os.Signal, 1)

	go func() {
		wApp.Slog.Info("Server running on port :8000")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			wApp.Slog.Error(err)
			gracefulShut <- syscall.SIGTERM
		}
	}()

	//Graceful shutdown
	wApp.graceful(s, app.Slog, gracefulShut)
}

func (wApp *WraperApp) graceful(hs *http.Server, slog *simplelog.Log, gracefulShut chan os.Signal) {

	signal.Notify(gracefulShut, os.Interrupt, syscall.SIGTERM)

	<-gracefulShut

	for client := range wApp.EventClients {
		delete(wApp.EventClients, client)
	}

	if err := hs.Shutdown(context.Background()); err != nil {
		slog.Error(err)
	} else {
		slog.Info("Server stopped")
	}
}

func LoadRoutes(wApp *WraperApp) http.Handler {
	wApp.AbsSrvRootDir, _ = filepath.Abs(wApp.SrvRootDir)
	fs := http.FileServer(http.Dir(wApp.AbsSrvRootDir))
	fmt.Println(wApp.AbsSrvRootDir)
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/event", wApp.handleEvent)
	mux.HandleFunc("/ws", wApp.handleConnections)
	mux.HandleFunc("/facebook_login", wApp.handleFacebookLogin)
	mux.HandleFunc("/facebookCallback", wApp.handleFacebookCallback)
	return mux
}

func (wApp *WraperApp) handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != wApp.FacebookState {
		wApp.Slog.Infof("invalid oauth state, expected '%s', got '%s'\n", wApp.FacebookState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := wApp.FacebookConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		wApp.Slog.Infof("oauthConf.Exchange() failed with '%s'\n", err)
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
	if !wApp.Db.Instance().UserExists(user.Email) {
		wApp.Db.Instance().AddNewUser(&user)

	}

	expiration := time.Now().Add(60 * time.Second)
	//Warning! used to set email in Cookie just for simplicity.
	cookie := http.Cookie{Name: "ticket_booking", Value: user.Email, Expires: expiration}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, r.Referer(), http.StatusTemporaryRedirect)
}

func (wApp *WraperApp) handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	authURL, err := url.Parse(wApp.FacebookConfig.Endpoint.AuthURL)
	if err != nil {
		wApp.Slog.Error(err)
	}
	parameters := url.Values{}
	parameters.Add("client_id", wApp.FacebookConfig.ClientID)
	parameters.Add("scope", strings.Join(wApp.FacebookConfig.Scopes, " "))
	parameters.Add("redirect_uri", wApp.FacebookConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", wApp.FacebookState)
	authURL.RawQuery = parameters.Encode()
	url := authURL.String()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (wApp *WraperApp) handleEvent(w http.ResponseWriter, r *http.Request) {
	tmplData := wApp.Db.Instance().AllPlacesInEvent()
	tmplData.Request = r

	cookie, ok := app.IsLogged(r)
	if !ok {
		tmplData.UserInfo["isLogged"] = "0"
	} else {
		user, _ := wApp.Db.Instance().FindUserByEmail(cookie.Value)
		tmplData.UserInfo["isLogged"] = "1"
		tmplData.UserInfo["name"] = user.Name
		tmplData.UserInfo["id"] = user.ID
	}

	app.PopulateTemplate(tmplData, w, fmt.Sprintf("%s/event.html", wApp.AbsSrvRootDir))
}

func (wApp *WraperApp) handleConnections(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		wApp.Slog.Fatal(err)
	}

	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		wApp.Slog.Error(err)
		return
	}

	event := string(msg)

	clients, ok := wApp.EventClients[event]

	if !ok {
		client := []*websocket.Conn{wsConn}
		wApp.EventClients[event] = client
	} else {
		clients = append(clients, wsConn)
		wApp.EventClients[event] = clients
	}

	// Defering here since we have defined eventClients for this particular connection
	defer func() {
		app.DeleteClient(wApp.EventClients[event], app.IndexOfClient(wsConn, wApp.EventClients[event]))
		wsConn.Close()
	}()

	for {
		var places models.EventPlaces
		//imagine data is not corrupted or mixed up
		err := wsConn.ReadJSON(&places)
		if err != nil {
			// wApp.Slog.Error(err)
			fmt.Println(err)
			return
		}

		wApp.Mu.Lock()
		code := wApp.Db.Instance().ProcessPlace(&places, app.BuildUserIdentity(wsConn.RemoteAddr().String()))
		wApp.Mu.Unlock()

		places.ErrorCode = code
		occupied := wApp.Db.Instance().OccupiedPlacesInEvent()

		for _, occupiedPlace := range occupied {
			if !app.InStringSlice(places.BookedPlaces, occupiedPlace.PlaceIdentity) {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
			}
		}

		places.UserAddr = wsConn.RemoteAddr().String()

		wApp.Broadcast <- places
	}
}

func (wApp *WraperApp) handlePlaceBookings() {
	for places := range wApp.Broadcast {
		for _, client := range wApp.EventClients[places.Event] {
			if client == nil {
				continue
			}
			if places.UserAddr == client.RemoteAddr().String() {
				err := client.WriteJSON(models.GetEventSystemMessage(places.ErrorCode, places.Event, places.LastActedPlace))

				if err != nil {
					wApp.Slog.Error(err)
				}
				continue
			}

			err := client.WriteJSON(places)
			fmt.Println(err)
			if err != nil {
				wApp.Slog.Error(err)
			}
		}
	}
}

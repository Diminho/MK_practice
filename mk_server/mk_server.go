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

	"io"

	"github.com/Diminho/MK_practice/app"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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

type Application interface {
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
			// TODO: context
			gracefulShut <- syscall.SIGTERM
		}
	}()

	signal.Notify(gracefulShut, os.Interrupt, syscall.SIGTERM)
	// select{gracefulShut, context}
	<-gracefulShut

	//Graceful shutdown
	wApp.graceful(s, app.Slog)
}

// TODO: Add context
func (wApp *WraperApp) graceful(hs *http.Server, slog *simplelog.Log) {
	for client := range wApp.EventClients {
		// TODO: All connects must be closed before exit
		delete(wApp.EventClients, client)
	}

	// TODO: Add timeout
	if err := hs.Shutdown(context.Background()); err != nil {
		slog.Error(err)
	}
	slog.Info("Server stopped")
}

func LoadRoutes(wApp *WraperApp) http.Handler {
	// TODO: Error
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
		wApp.Slog.Info(errors.Wrap(err, "oauthConf.Exchange() failed"))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://graph.facebook.com/me?fields=email,name,id&access_token=" +
		url.QueryEscape(token.AccessToken))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer func() {
		// TODO: Error
		io.Copy(ioutil.Discard, resp.Body)
		err := resp.Body.Close()
		if err != nil {
			wApp.Slog.Error(err)
		}
	}()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	var user models.User
	// TODO: Error
	json.Unmarshal(response, &user)

	if !wApp.Db.Instance().UserExists(user.Email) {
		// TODO: ERROR
		wApp.Db.Instance().AddNewUser(&user)
	}

	expiration := time.Now().Add(60 * time.Second)
	//Warning! used to set email in Cookie just for simplicity.
	// TODO: implement User Side Session store
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

	if clients, ok := wApp.EventClients[event]; !ok {
		client := []*websocket.Conn{wsConn}
		wApp.EventClients[event] = client
	} else {
		clients = append(clients, wsConn)
		wApp.EventClients[event] = clients
	}

	// Defering here since we have defined eventClients for this particular connection
	defer func() {
		// TODO: Use User ID to delete from EventClients
		app.DeleteClient(wApp.EventClients[event], app.IndexOfClient(wsConn, wApp.EventClients[event]))
		// TODO: Error
		wsConn.Close()
	}()

	for {
		var places models.EventPlaces
		//imagine data is not corrupted or mixed up
		err := wsConn.ReadJSON(&places)
		if err != nil {
			// TODO: Handle EOF separately
			wApp.Slog.Error(err)
			//fmt.Println(err)
			return
		}

		// TODO: Split event locking by event ID
		wApp.Mu.Lock()
		// TODO: User ID
		places.ErrorCode = wApp.Db.Instance().ProcessPlace(&places, app.BuildUserIdentity(wsConn.RemoteAddr().String()))
		wApp.Mu.Unlock()

		occupied := wApp.Db.Instance().OccupiedPlacesInEvent()

		existing := make(map[string]struct{})
		for _, occupiedPlace := range occupied {
			if _, ok := existing[occupiedPlace.PlaceIdentity]; !ok {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
				existing[occupiedPlace.PlaceIdentity] = struct{}{}
			}
		}

		// TODO: Do not share user IP
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

			if places.UserAddr != client.RemoteAddr().String() {
				err := client.WriteJSON(places)
				if err != nil {
					wApp.Slog.Error(err)
				}
				continue
			}

			err := client.WriteJSON(models.GetEventSystemMessage(places.ErrorCode, places.Event, places.LastActedPlace))

			if err != nil {
				wApp.Slog.Error(err)
			}
		}
	}
}

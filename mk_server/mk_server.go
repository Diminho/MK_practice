package mk_server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
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

func NewServer(ctx context.Context, app *app.App) {
	ctx, cancel := context.WithCancel(ctx)
	wApp := &WraperApp{app}

	s := &http.Server{Addr: ":8000", Handler: LoadRoutes(wApp)}

	go wApp.handlePlaceBookings()

	go func() {
		wApp.Slog.Info("Server running on port :8000")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			wApp.Slog.Error(err)
		}
		cancel()
	}()
	var gracefulShut = make(chan os.Signal, 1)
	signal.Notify(gracefulShut, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-gracefulShut:
	}

	//Graceful shutdown
	wApp.graceful(ctx, s, app.Slog)
}

func (wApp *WraperApp) graceful(ctx context.Context, hs *http.Server, slog *simplelog.Log) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, events := range wApp.EventClients {
		for _, client := range events {
			if err := client.Close(); err != nil {
				slog.Error(err)
			}
		}
	}

	if err := hs.Shutdown(ctx); err != nil {
		slog.Error(err)
	}
	close(wApp.Broadcast)
	// app.RemoveContents("./tmp/")
	app.RemoveContents(os.TempDir())

	slog.Info("Server stopped")
}

func LoadRoutes(wApp *WraperApp) http.Handler {
	absSrvRootDir, err := filepath.Abs(wApp.SrvRootDir)
	if err != nil {
		wApp.Slog.Error(err)
	}
	//need this assignemt since wApp.AbsSrvRootDir is used in application
	wApp.AbsSrvRootDir = absSrvRootDir
	fs := http.FileServer(http.Dir(wApp.AbsSrvRootDir))
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
		wApp.Slog.Info(fmt.Sprintf("invalid oauth state, expected '%s', got '%s'\n", wApp.FacebookState, state))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")

	token, err := wApp.FacebookConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		wApp.Slog.Info(fmt.Sprintf("oauthConf.Exchange() failed with '%s'\n", err))
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
		_, err := io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			wApp.Slog.Error(err)
		}

		err = resp.Body.Close()
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
	err = json.Unmarshal(response, &user)
	if err != nil {
		wApp.Slog.Error(err)
	}

	userExists, err := wApp.Db().UserExists(user.Email)
	if err != nil {
		wApp.Slog.Error(err)
	}
	if !userExists {
		err := wApp.Db().AddNewUser(&user)
		if err != nil {
			wApp.Slog.Error(err)
		}
	}

	session, err := wApp.Manager.SessionStart(w, r)
	if err != nil {
		wApp.Slog.Error(err)
	}
	app.AuthUser(r, session, &user)

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
	tmplData, err := wApp.Db().AllPlacesInEvent()

	if err != nil {
		wApp.Slog.Error(err)
	}
	session, err := wApp.Manager.SessionStart(w, r)

	if err != nil {
		wApp.Slog.Error(err)
	}
	ok, err := app.IsLogged(r, session)
	if err != nil {
		wApp.Slog.Error(err)
	}
	if !ok {
		tmplData.UserInfo["isLogged"] = "0"
	} else {
		email, err := session.Get("email")
		if err != nil {
			wApp.Slog.Error(err)
		}

		user, err := wApp.Db().FindUserByEmail(email.(string))
		if err != nil {
			wApp.Slog.Error(err)
		}
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
		app.DeleteClient(wApp.EventClients[event], app.IndexOfClient(wsConn, wApp.EventClients[event]))

		if err := wsConn.Close(); err != nil {
			wApp.Slog.Error(err)
		}
	}()

	for {
		places := models.EventPlaces{Mutex: &sync.Mutex{}}
		//imagine data is not corrupted or mixed up
		err := wsConn.ReadJSON(&places)
		if err != nil {
			// we use io.ErrUnexpectedEOF instead of io.EOF since ReadJSON returns io.ErrUnexpectedEOF in case of io.EOF
			if err == io.ErrUnexpectedEOF {
				break
			}
			wApp.Slog.Error(err)
			return
		}

		places.Lock()
		places.ErrorCode, err = wApp.Db().ProcessPlace(&places, app.BuildUserIdentity(wsConn.RemoteAddr().String()))
		places.Unlock()

		if err != nil {
			wApp.Slog.Error(err)
		}

		occupied, err := wApp.Db().OccupiedPlacesInEvent()

		if err != nil {
			wApp.Slog.Error(err)
		}

		existing := make(map[string]struct{})
		for _, occupiedPlace := range occupied {
			if _, ok := existing[occupiedPlace.PlaceIdentity]; !ok {
				places.BookedPlaces = append(places.BookedPlaces, occupiedPlace.PlaceIdentity)
				existing[occupiedPlace.PlaceIdentity] = struct{}{}
			}
		}

		places.UserAddr = wsConn.RemoteAddr().String()

		wApp.Broadcast <- places
	}
}

func (wApp *WraperApp) handlePlaceBookings() {
	for places := range wApp.Broadcast {
		fmt.Println(wApp.EventClients[places.Event])
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

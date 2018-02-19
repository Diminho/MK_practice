package app

import (
	"github.com/Diminho/MK_practice/mk_session"

	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

type App struct {
	Broadcast      chan models.EventPlaces
	EventClients   map[string][]*websocket.Conn
	FacebookState  string
	FacebookConfig *oauth2.Config
	Slog           *simplelog.Log
	Db             func() models.Database
	SrvRootDir     string
	AbsSrvRootDir  string
	Manager        *mk_session.Manager
}

func NewApp(slog *simplelog.Log, fbConfigOauth *oauth2.Config, FBState string, SrvRootDir string, instance func() models.Database) *App {
	app := &App{
		EventClients:   make(map[string][]*websocket.Conn),
		Broadcast:      make(chan models.EventPlaces, 1),
		FacebookConfig: fbConfigOauth,
		FacebookState:  FBState,
		Db:             instance,
		Slog:           slog,
		SrvRootDir:     SrvRootDir,
		Manager:        mk_session.NewManager("sessionID", 60),
	}

	return app
}

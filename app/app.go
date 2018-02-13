package app

import (
	"sync"

	"github.com/Diminho/MK_practice/config"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

// TODO: Hide internal fields
type App struct {
	Broadcast      chan models.EventPlaces
	EventClients   map[string][]*websocket.Conn
	FacebookState  string
	FacebookConfig *oauth2.Config
	Slog           *simplelog.Log
	Mu             *sync.Mutex
	Db             models.Database
	SrvRootDir     string
	AbsSrvRootDir  string
}

// TODO: Should be configurable
func InitApp(db models.Database, slog *simplelog.Log) *App {
	app := &App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces, 1),
		FacebookConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     config.Endpoint,
		},
		FacebookState: "MK_PRACTICE",
		Db:            db,
		Slog:          slog,
		Mu:            &sync.Mutex{},
		SrvRootDir:    "./public",
	}

	return app
}

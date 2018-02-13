package app

import (
	"sync"

	"github.com/Diminho/MK_practice/config"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"
	"golang.org/x/oauth2"
)

// type App struct {
// 	Broadcast      chan models.EventPlaces
// 	EventClients   map[string][]*websocket.Conn
// 	FacebookState  string
// 	FacebookConfig *oauth2.Config
// 	Slog           *simplelog.Log
// 	Mu             *sync.Mutex
// 	Db             models.Database
// }

// func InitApp(db *models.DB, slog *simplelog.Log) *App {
// 	app := &App{
// 		EventClients: make(map[string][]*websocket.Conn),
// 		Broadcast:    make(chan models.EventPlaces),
// 		FacebookConfig: &oauth2.Config{
// 			ClientID:     config.ClientID,
// 			ClientSecret: config.ClientSecret,
// 			RedirectURL:  config.RedirectURL,
// 			Scopes:       config.Scopes,
// 			Endpoint:     config.Endpoint,
// 		},
// 		FacebookState: "MK_PRACTICE",
// 		Db:            db,
// 		Slog:          slog,
// 		Mu:            &sync.Mutex{},
// 	}

// 	return app
// }

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

func InitApp(db *models.DB, slog *simplelog.Log) *App {
	app := &App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces),
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

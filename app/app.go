package app

import (
	"html/template"
	"path/filepath"

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
	Tmpl           *template.Template
}

func NewApp(slog *simplelog.Log, fbConfigOauth *oauth2.Config, FBState string, SrvRootDir string, instance func() models.Database) (*App, error) {

	absSrvRootDir, err := filepath.Abs(SrvRootDir)

	if err != nil {
		slog.Error(err)
	}

	tmpl, err := ParseTemplates(absSrvRootDir)

	if err != nil {
		return nil, err
	}

	app := &App{
		EventClients:   make(map[string][]*websocket.Conn),
		Broadcast:      make(chan models.EventPlaces, 1),
		FacebookConfig: fbConfigOauth,
		FacebookState:  FBState,
		Db:             instance,
		Slog:           slog,
		SrvRootDir:     SrvRootDir,
		Manager:        mk_session.NewManager("sessionID", 60),
		AbsSrvRootDir:  absSrvRootDir,
		Tmpl:           tmpl,
	}

	return app, err
}

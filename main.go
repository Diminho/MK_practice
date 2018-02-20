package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Diminho/MK_practice/app"
	"github.com/Diminho/MK_practice/config"
	"github.com/Diminho/MK_practice/mk_server"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
)

func main() {
	// os.RemoveAll("/tmp/")
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	slog := simplelog.NewLog(file)

	db, err := models.Connect(config.Driver, fmt.Sprintf("%s:@%s/%s", config.User, config.Host, config.DbName))
	if err != nil {
		slog.Fatal(err)
	}

	mk_server.NewServer(
		app.NewApp(
			slog,
			&oauth2.Config{
				ClientID:     config.ClientID,
				ClientSecret: config.ClientSecret,
				RedirectURL:  config.RedirectURL,
				Scopes:       config.Scopes,
				Endpoint:     config.Endpoint,
			},
			config.FBState,
			config.ServerRootDir,
			db.Instance()))
}

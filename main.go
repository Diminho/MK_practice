package main

import (
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Diminho/MK_practice/app"
	"github.com/Diminho/MK_practice/mk_server"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	logjson "github.com/Diminho/MK_practice/simplelog/handlers/json"
)

func main() {

	file, fileErr := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if fileErr != nil {
		log.Fatal("Failed to open log file: ", fileErr)
	}
	defer file.Close()

	slog := simplelog.NewLog(file)
	slog.SetHandler(logjson.New(slog))

	db, dbErr := models.Connect()
	if dbErr != nil {
		slog.Fatal(dbErr)
	}

	app := app.InitApp(db, slog)

	mk_server.NewServer(app)
}

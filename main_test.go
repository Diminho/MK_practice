package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Diminho/MK_practice/models"
	"github.com/gorilla/websocket"
)

var app *App

type mockDB struct{}

func (mdb *mockDB) AllPlacesInEvent() models.EventPlacesTemplate {

	templateRows := models.EventPlacesTemplate{}
	templateRows.UserInfo = make(map[string]string)
	eventPlacesRows := []models.EventPlacesRow{}
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat1", 1, 1, "user1"})
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat2", 0, 0, "user2"})
	templateRows.EventPlacesRows = eventPlacesRows

	return templateRows
}

func init() {
	app = &App{
		eventClients: make(map[string][]*websocket.Conn),
		broadcast:    make(chan models.EventPlaces),
	}
}

func testHandleEvent(t *testing.T) {

	req, err := http.NewRequest("GET", "/event", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	app.handleEvent(rec, req)
}

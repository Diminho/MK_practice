package app_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Diminho/MK_practice/app"
	"github.com/Diminho/MK_practice/mk_server"
	"github.com/Diminho/MK_practice/models"
	"github.com/gorilla/websocket"
)

var testApp *mk_server.WraperApp

func init() {
	testApp = &mk_server.WraperApp{&app.App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces),
		SrvRootDir:   "../public",
		Mu:           &sync.Mutex{},
	}}

}

func TestInStringSliceTrue(t *testing.T) {
	input := []string{"first", "second", "third"}
	needle := "third"
	if !app.InStringSlice(input, needle) {
		t.Errorf("needle [%s] is not in slice %v", needle, input)
	}
}

func TestInStringSliceFalse(t *testing.T) {
	input := []string{"first", "second", "third"}
	needle := "fourth"
	if app.InStringSlice(input, needle) {
		t.Errorf("needle [%s] should not be in slice %v", needle, input)
	}
}

func TestIsLogged(t *testing.T) {
	srv := httptest.NewServer(mk_server.LoadRoutes(testApp))
	defer srv.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/event", srv.URL), nil)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	req.AddCookie(&http.Cookie{Name: "ticket_booking", Value: "user_1"})

	_, ok := app.IsLogged(req)
	if !ok {
		t.Error("user is not logged")
	}
}

func TestBuildUserIdentity(t *testing.T) {
	addr := "10.11.12.13:1111"
	gotUserID := app.BuildUserIdentity(addr)
	expectedUserID := "10.11.12.13_1111"

	if expectedUserID != gotUserID {
		t.Errorf("expected %s, got %s", expectedUserID, gotUserID)
	}
}

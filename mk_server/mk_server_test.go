package mk_server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"

	"github.com/Diminho/MK_practice/models"
	"github.com/gorilla/websocket"

	"github.com/Diminho/MK_practice/app"
)

var testApp *WraperApp

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

func (mdb *mockDB) AddNewUser(user *models.User) error {
	return nil
}

func (mdb *mockDB) FindUserByEmail(email string) (models.User, error) {
	user := models.User{Name: "John", Email: "test@email.com", FacebookID: "000000", ID: "0007"}

	return user, nil
}

func (mdb *mockDB) Instance() models.Database {
	return mdb
}

func (mdb *mockDB) OccupiedPlacesInEvent() []models.EventPlacesRow {
	return nil
}

func (mdb *mockDB) ProcessPlace(places *models.EventPlaces, user string) int {
	return 0
}

func (mdb *mockDB) UserExists(email string) bool {
	return false
}

// TODO: Should not be common
func init() {
	testApp = &WraperApp{&app.App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces),
		Db:           &mockDB{},
		SrvRootDir:   "../public",
		Mu:           &sync.Mutex{},
	}}
}

func TestHandleEventWithCookies(t *testing.T) {
	// testify/assert

	srv := httptest.NewServer(LoadRoutes(testApp))

	defer srv.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/event", srv.URL), nil)

	req.AddCookie(&http.Cookie{Name: "ticket_booking", Value: "user_1"})

	var client = &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK %s", res.Status)
	}
}

func TestHandleEvent(t *testing.T) {
	// t.Run()

	srv := httptest.NewServer(LoadRoutes(testApp))

	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/event", srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK %v", res.Status)
	}

}

func TestHandleConnections(t *testing.T) {
	srv := httptest.NewServer(LoadRoutes(testApp))
	defer srv.Close()

	go testApp.handlePlaceBookings()

	u, _ := url.Parse(fmt.Sprintf("%s/ws", srv.URL))
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("cannot make websocket connection: %v", err)
	}

	//first send to set event name
	err = conn.WriteMessage(websocket.BinaryMessage, []byte(`test_event`))
	if err != nil {
		t.Fatalf("cannot write message: %v", err)
	}

	// var places models.EventPlaces
	var gotEventSysMsg models.EventSystemMessage

	input := models.EventPlaces{Event: "test_event", BookedPlaces: []string{"seat_6"}, Action: "book", LastActedPlace: "seat_6"}

	err = conn.WriteJSON(input)
	if err != nil {
		t.Errorf("cannot write message: %v", err)
	}
	err = conn.ReadJSON(&gotEventSysMsg)

	if err != nil {
		t.Errorf("cannot read message: %v", err)
	}

	expected := models.EventSystemMessage{Message: "Success", MessageType: 0, Event: "test_event", LastActedPlace: "seat_6", BookTime: 60}

	if !reflect.DeepEqual(expected, gotEventSysMsg) {
		fmt.Println("equal")
		t.Errorf("expected %v , got %v", expected, gotEventSysMsg)
	}

	// TODO: Error
	conn.Close()
}
func TestHandleConnectionsWithOtherClients(t *testing.T) {
	//setting up server
	srv := httptest.NewServer(LoadRoutes(testApp))
	defer srv.Close()

	go testApp.handlePlaceBookings()

	u, _ := url.Parse(fmt.Sprintf("%s/ws", srv.URL))
	u.Scheme = "ws"

	var wg sync.WaitGroup

	testInput := models.EventPlaces{Event: "test_event", BookedPlaces: []string{"seat_5"}, Action: "book", LastActedPlace: "seat_5"}

	wg.Add(2)
	// FIRST CLIENT
	go func() {
		defer wg.Done()
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("cannot make websocket connection: %v", err)
		}

		//first send to set event name
		err = conn.WriteMessage(websocket.BinaryMessage, []byte(`test_event`))
		if err != nil {
			t.Errorf("cannot write message: %v", err)
		}

		// var places models.EventPlaces
		var gotEventSysMsg models.EventSystemMessage

		err = conn.WriteJSON(testInput)
		if err != nil {
			t.Errorf("cannot write message: %v", err)
		}
		err = conn.ReadJSON(&gotEventSysMsg)

		if err != nil {
			t.Errorf("cannot read message: %v", err)
		}

		expected := models.EventSystemMessage{Message: "Success", MessageType: 0, Event: "test_event", LastActedPlace: "seat_5", BookTime: 60}

		if !reflect.DeepEqual(expected, gotEventSysMsg) {
			fmt.Println("equal")
			t.Errorf("expected %v , got %v", expected, gotEventSysMsg)
		}
		conn.Close()
	}()

	// SECOND CLIENT
	go func() {
		defer wg.Done()
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("cannot make websocket connection: %v", err)
		}

		//first send to set event name
		err = conn.WriteMessage(websocket.BinaryMessage, []byte(`test_event`))
		if err != nil {
			t.Errorf("cannot write message: %v", err)
		}

		var got models.EventPlaces
		err = conn.ReadJSON(&got)
		if err != nil {
			t.Error("cannot read message: ", err)
		}
		t.Log("success: received response: ", got)

		// we dont want to comtare user remote address
		got.UserAddr = ""
		if !reflect.DeepEqual(testInput, got) {
			t.Errorf("expected %v , got %v", testInput, got)
		}
		conn.Close()
	}()

	wg.Wait()

}

// TODO: Add some negative cases
// TODO: Try testify or https://github.com/matryer/is
// TODO: Try t.Run()

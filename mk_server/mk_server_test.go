package mk_server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/Diminho/MK_practice/mk_session"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"

	"github.com/Diminho/MK_practice/app"
	"github.com/stretchr/testify/assert"
)

var testApp *WraperApp

type mockDB struct{}

func (mdb *mockDB) AllPlacesInEvent() (models.EventPlacesTemplate, error) {

	templateRows := models.EventPlacesTemplate{}
	templateRows.UserInfo = make(map[string]string)
	eventPlacesRows := []models.EventPlacesRow{}
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat1", 1, 1, "user1"})
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat2", 0, 0, "user2"})
	templateRows.EventPlacesRows = eventPlacesRows

	return templateRows, nil
}

func (mdb *mockDB) AddNewUser(user *models.User) error {
	return nil
}

func (mdb *mockDB) FindUserByEmail(email string) (models.User, error) {
	user := models.User{Name: "John", Email: "test@email.com", FacebookID: "000000", ID: "0007"}

	return user, nil
}

func (mdb *mockDB) Instance() func() models.Database {

	return func() models.Database {
		return mdb
	}
}

func (mdb *mockDB) OccupiedPlacesInEvent() ([]models.EventPlacesRow, error) {
	return nil, nil
}

func (mdb *mockDB) ProcessPlace(places *models.EventPlaces, user string) (models.StatusCode, error) {
	return 0, nil
}

func (mdb *mockDB) UserExists(email string) (bool, error) {
	return false, nil
}

func init() {

	file, err := os.OpenFile("log_test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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

	testApp = &WraperApp{&app.App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces),
		Db:           (&mockDB{}).Instance(),
		SrvRootDir:   "../public",
		Manager:      mk_session.NewManager("sessionTestID", 30),
		Slog:         slog,
	}}

}

func TestHandleEventWithCookies(t *testing.T) {

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

	srv := httptest.NewServer(LoadRoutes(testApp))

	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/event", srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, http.StatusOK, "StatusCode is not 200", res.Status)
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
		t.Fatal("cannot write message: %v", err)
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

		input := models.EventPlaces{Event: "test_event", BookedPlaces: []string{"seat_5"}, Action: "book", LastActedPlace: "seat_5"}

		err = conn.WriteJSON(input)
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
			t.Errorf("cannot read message: %v", err)
		}
		fmt.Printf("success: received response: %v\n", got)

		expected := models.EventPlaces{BookedPlaces: []string{"seat_5"}, Event: "test_event", LastActedPlace: "seat_5", Action: "book"}
		// we dont want to comtare user remote address
		got.UserAddr = ""
		if !reflect.DeepEqual(expected, got) {
			fmt.Println("equal")
			t.Errorf("expected %v , got %v", expected, got)
		}
		conn.Close()
	}()

	wg.Wait()

}

package mk_server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Diminho/MK_practice/mk_session"
	"github.com/Diminho/MK_practice/models"
	"github.com/Diminho/MK_practice/simplelog"
	"github.com/gorilla/websocket"

	"github.com/Diminho/MK_practice/app"
	"github.com/stretchr/testify/assert"
)

var testApp *WraperApp

type mockDB struct {
	fieldToTestNegativeCase bool
}

func (mdb *mockDB) AllPlacesInEvent() (models.EventPlacesTemplate, error) {

	templateRows := models.EventPlacesTemplate{}
	templateRows.UserInfo = make(map[string]string)
	eventPlacesRows := []models.EventPlacesRow{}
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat1", 1, 1, "user1"})
	eventPlacesRows = append(eventPlacesRows, models.EventPlacesRow{"seat2", 0, 0, "user2"})
	templateRows.EventPlacesRows = eventPlacesRows

	if mdb.fieldToTestNegativeCase {
		return templateRows, errors.New("testerror")
	}
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

func initApp(out io.Writer, negativeCase bool) *WraperApp {

	slog := simplelog.NewLog(out)

	testApp = &WraperApp{&app.App{
		EventClients: make(map[string][]*websocket.Conn),
		Broadcast:    make(chan models.EventPlaces),
		Db:           (&mockDB{fieldToTestNegativeCase: negativeCase}).Instance(),
		SrvRootDir:   "../public",
		Manager:      mk_session.NewManager("sessionTestID", 5),
		Slog:         slog,
	}}
	return testApp
}

func TestHandleEvent(t *testing.T) {
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

	t.Run("Testing just response", func(t *testing.T) {

		srv := httptest.NewServer(LoadRoutes(initApp(file, false)))

		defer srv.Close()

		res, err := http.Get(fmt.Sprintf("%s/event", srv.URL))
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, res.StatusCode, http.StatusOK, "StatusCode is not 200", res.Status)
	})
	t.Run("Testing just response- negative", func(t *testing.T) {

		srv := httptest.NewServer(LoadRoutes(initApp(file, true)))

		defer srv.Close()

		res, err := http.Get(fmt.Sprintf("%s/event", srv.URL))
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, res.StatusCode, http.StatusOK, "StatusCode is not 200", res.Status)
	})
	//wait until session files will be deleted
	time.Sleep(5 * time.Second)
}

func TestHandleConnections(t *testing.T) {

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

	srv := httptest.NewServer(LoadRoutes(initApp(file, false)))
	defer srv.Close()

	go testApp.handlePlaceBookings()

	u, _ := url.Parse(fmt.Sprintf("%s/ws", srv.URL))
	u.Scheme = "ws"

	t.Run("test_self_connection", func(t *testing.T) {
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

		assert.Equal(t, expected, gotEventSysMsg, "should be equal")

		conn.Close()
	})

	t.Run("test connection with clients", func(t *testing.T) {
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

			assert.Equal(t, expected, gotEventSysMsg, "should be equal")
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

			assert.Equal(t, expected, got, "should be equal")
			conn.Close()
		}()

		wg.Wait()
	})
}

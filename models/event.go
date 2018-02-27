package models

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

const BookTime int = 60

type StatusCode int

type EventPlaces struct {
	*sync.Mutex
	Event          string     `json:"event"`
	BookedPlaces   []string   `json:"places"`
	LastActedPlace string     `json:"lastActedPlace"`
	Action         string     `json:"action"`
	UserAddr       string     `json:"-"`
	ErrorCode      StatusCode `json:"errorCode"`
}

type EventPlacesRow struct {
	PlaceIdentity string
	IsBooked      int
	IsBought      int
	UserID        string
}

type EventPlacesTemplate struct {
	EventPlacesRows []EventPlacesRow
	EventID         int
	UserInfo        map[string]string
}

type EventSystemMessage struct {
	Message        string     `json:"sysMessage"`
	MessageType    StatusCode `json:"messageType"`
	Event          string     `json:"event"`
	LastActedPlace string     `json:"lastActedPlace"`
	BookTime       int        `json:"bookTime"`
}

func (db *DB) OccupiedPlacesInEvent() (eventRow []EventPlacesRow, err error) {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places where isBooked = 1 OR isBought = 1"
	event, err := queryEvent(sqlStatement, db)
	eventRow = event.EventPlacesRows
	return
}

func (db *DB) AllPlacesInEvent() (event EventPlacesTemplate, err error) {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places"
	event, err = queryEvent(sqlStatement, db)
	event.UserInfo = make(map[string]string)
	return
}

func queryEvent(sqlStatement string, db *DB) (EventPlacesTemplate, error) {
	templateRows := EventPlacesTemplate{}
	rows, err := db.Query(sqlStatement)

	if err != nil {
		return templateRows, err
	}

	defer rows.Close()

	for rows.Next() {
		eventPlacesRow := EventPlacesRow{}
		err := rows.Scan(&eventPlacesRow.PlaceIdentity, &eventPlacesRow.IsBooked, &eventPlacesRow.IsBought, &eventPlacesRow.UserID, &templateRows.EventID)

		if err != nil {
			return templateRows, err
		}

		templateRows.EventPlacesRows = append(templateRows.EventPlacesRows, eventPlacesRow)
	}

	err = rows.Err()

	if err != nil {
		log.Fatal(err)
		return templateRows, err
	}

	return templateRows, nil
}

func (db *DB) ProcessPlace(places *EventPlaces, user string) (StatusCode, error) {
	var sqlStatement string

	switch places.Action {
	case "buy":
		sqlStatement = "UPDATE event_places set isBought = 1, isBooked = 0 WHERE placeIdentity = ?"
	case "rejected_timeout":

		args := make([]interface{}, len(places.BookedPlaces))
		for i, id := range places.BookedPlaces {
			args[i] = id
		}

		sqlStatement = "UPDATE event_places set isBooked = 0, userId = ''  WHERE placeIdentity IN(?" + strings.Repeat(",?", len(args)-1) + ")"
		stmt, err := db.Prepare(sqlStatement)

		if err != nil {
			return 1, err
		}
		_, err = stmt.Exec(args...)

		if err != nil {
			return 1, err
		}
		places.BookedPlaces = []string{}

		return 0, nil
	case "unbook":
		sqlStatement = "UPDATE event_places set isBooked = 0, userId = '' WHERE placeIdentity = ?"
	case "book":
		var isBooked int
		err := db.QueryRow("SELECT isBooked FROM event_places WHERE placeIdentity = ?", places.LastActedPlace).Scan(&isBooked)
		if err != nil {
			return 1, err
		}

		if isBooked == 1 {
			//1 - code error - is a booked earlier by other user
			return 1, nil
		}
		sqlStatement = "UPDATE event_places set isBooked = 1, userId = '" + user + "' WHERE placeIdentity = ?"
	}

	stmt, err := db.Prepare(sqlStatement)

	if err != nil {
		return 1, err
	}

	_, err = stmt.Exec(places.LastActedPlace)

	if err != nil {
		return 1, err
	}

	return 0, nil
}

func GetEventSystemMessage(code StatusCode, event string, place string) EventSystemMessage {
	var message string

	switch code {
	case 0:
		message = "Success"
	case 1:
		message = fmt.Sprintf("Sorry, this [%s] have been already booked", place)
	}

	return EventSystemMessage{Message: message, MessageType: code, Event: event, LastActedPlace: place, BookTime: BookTime}
}

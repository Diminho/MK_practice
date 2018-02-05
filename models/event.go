package models

import (
	"log"
	"net/http"
	"strings"
)

const BookTime int = 60

type EventPlaces struct {
	Event          string   `json:"event"`
	BookedPlaces   []string `json:"places"`
	LastActedPlace string   `json:"lastActedPlace"`
	Action         string   `json:"action"`
	UserAddr       string   `json:"userAddr"`
	ErrorCode      int      `json:"errorCode"`
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
	Request         *http.Request
	UserInfo        map[string]string
}

func (db *DB) OccupiedPlacesInEvent() []EventPlacesRow {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places where isBooked = 1 OR isBought = 1"
	event := queryEvent(sqlStatement, db)

	return event.EventPlacesRows
}

func (db *DB) AllPlacesInEvent() EventPlacesTemplate {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places"
	event := queryEvent(sqlStatement, db)
	event.UserInfo = make(map[string]string)
	return event
}

func queryEvent(sqlStatement string, db *DB) EventPlacesTemplate {
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatal(err)
	}

	templateRows := EventPlacesTemplate{}
	defer rows.Close()

	for rows.Next() {
		eventPlacesRow := EventPlacesRow{}
		err := rows.Scan(&eventPlacesRow.PlaceIdentity, &eventPlacesRow.IsBooked, &eventPlacesRow.IsBought, &eventPlacesRow.UserID, &templateRows.EventID)

		if err != nil {
			log.Fatal(err)
		}

		templateRows.EventPlacesRows = append(templateRows.EventPlacesRows, eventPlacesRow)
	}

	err = rows.Err()

	if err != nil {
		log.Fatal(err)
	}

	return templateRows
}

func (db *DB) ProcessPlace(places *EventPlaces, user string) int {
	var sqlStatement string
	var isBooked int

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
			log.Fatal(err)
		}
		_, err = stmt.Exec(args...)

		if err != nil {
			log.Fatal(err)
		}
		places.BookedPlaces = []string{}

		return 0
	case "unbook":
		sqlStatement = "UPDATE event_places set isBooked = 0, userId = '' WHERE placeIdentity = ?"
	case "book":
		err := db.QueryRow("SELECT isBooked FROM event_places WHERE placeIdentity = ?", places.LastActedPlace).Scan(&isBooked)
		if err != nil {
			log.Fatal(err)
		}

		if isBooked == 1 {
			//1 - code error - is a booked earlier by other user
			return 1
		}
		sqlStatement = "UPDATE event_places set isBooked = 1, userId = '" + user + "' WHERE placeIdentity = ?"
	}

	stmt, err := db.Prepare(sqlStatement)

	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(places.LastActedPlace)

	if err != nil {
		log.Fatal(err)
	}

	return 0
}

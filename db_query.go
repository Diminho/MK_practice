package main

import "log"

func queryForOccupiedPlacesInEvent() []EventPlacesRow {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places where isBooked = 1 OR isBought = 1"
	event := queryEvent(sqlStatement)

	return event.EventPlacesRows
}

func queryForAllPlacesInEvent() EventPlacesTemplate {
	sqlStatement := "select placeIdentity, isBooked, isBought, userId, eventId from event_places"
	event := queryEvent(sqlStatement)

	return event
}

func queryEvent(sqlStatement string) EventPlacesTemplate {
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

func processPlace(placeID string, action string, user string) int {
	var sqlStatement string
	var isBooked int

	switch action {
	case "buy":
		sqlStatement = "UPDATE event_places set isBought = 1 WHERE placeIdentity = ?"
	case "unbook":
		sqlStatement = "UPDATE event_places set isBooked = 0 WHERE placeIdentity = ?"
	case "book":
		err := db.QueryRow("SELECT isBooked FROM event_places WHERE placeIdentity = ?", placeID).Scan(&isBooked)
		if err != nil {
			log.Fatal(err)
		}

		if isBooked == 1 {
			//1 - code error - is a booked earlier by other user
			return 1
		}
		sqlStatement = "UPDATE event_places set isBooked = 1, userId = '" + user[6:] + "' WHERE placeIdentity = ?"
	}

	stmt, err := db.Prepare(sqlStatement)

	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(placeID)

	if err != nil {
		log.Fatal(err)
	}

	return 0
}

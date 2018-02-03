package models

import (
	"database/sql"

	"fmt"
)

type Database interface {
	QueryForOccupiedPlacesInEvent() []EventPlacesRow
	QueryForAllPlacesInEvent() EventPlacesTemplate
	ProcessPlace(*EventPlaces, string) int
	CheckIfUserExists(string) bool
	AddNewUser(*User) error
	FindUserByEmail(string) (User, error)
}

type DB struct {
	*sql.DB
}

func Connect(user string, host string, dbName string) (*DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:@%s/%s", user, host, dbName))

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

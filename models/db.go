package models

import (
	"database/sql"
	"log"
	"time"

	"fmt"

	"github.com/Diminho/MK_practice/config"
)

type Database interface {
	OccupiedPlacesInEvent() []EventPlacesRow
	AllPlacesInEvent() EventPlacesTemplate
	ProcessPlace(*EventPlaces, string) int
	UserExists(string) bool
	AddNewUser(*User) error
	FindUserByEmail(string) (User, error)
	Instance() Database
}

type DB struct {
	*sql.DB
}

func Connect() (*DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:@%s/%s", config.User, config.Host, config.DbName))

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Instance() Database {

	var connected bool
	connected, err := db.isAlive()
	if err != nil {
		log.Println("isAlive error: ", err)
	}

	for connected != true { // reconnect if we lost connection
		log.Print("Connection to MySQL was lost. Waiting for 3s...")
		db.Close()
		time.Sleep(3 * time.Second)
		log.Print("Reconnecting...")
		db, err = Connect()
		if err != nil {
			log.Println("Coonect error: ", err)
		}
		connected, err = db.isAlive()
		if err != nil {
			log.Println("isAlive error: ", err)
		}
	}

	return db

}

func (db *DB) isAlive() (bool, error) {
	if err := db.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

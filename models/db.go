package models

import (
	"database/sql"
	"time"
)

const retriesNumber int = 3

type Database interface {
	OccupiedPlacesInEvent() []EventPlacesRow
	AllPlacesInEvent() EventPlacesTemplate
	ProcessPlace(*EventPlaces, string) StatusCode
	UserExists(string) (bool, error)
	AddNewUser(*User) error
	FindUserByEmail(string) (User, error)
	// Instance() Database
	Instance() func() Database
}

type DB struct {
	*sql.DB
	driver string
	dsn    string
}

func Connect(driver, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, driver: driver, dsn: dataSourceName}, nil
}

func (db *DB) Instance() func() Database {
	return func() Database {
		var connected bool
		connected, err := db.isAlive()
		if err != nil {
			panic(err)
		}

		retriesLeft := retriesNumber

		for connected != true { // reconnect if we lost connection
			err := db.Close()
			if err != nil {
				panic(err)
			}
			time.Sleep(3 * time.Second)
			db, err = Connect(db.driver, db.dsn)
			if err != nil {
				panic(err)
			}
			connected, err = db.isAlive()
			if err != nil {
				panic(err)
			}
			if retriesLeft == 0 {
				err := db.Close()
				if err != nil {
					panic(err)
				}
				break
			}
			retriesLeft--
		}

		return db
	}

}

func (db *DB) isAlive() (bool, error) {
	if err := db.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

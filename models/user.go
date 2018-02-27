package models

import (
	"database/sql"
	"log"
)

type User struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	FacebookID string `json:"id"`
	ID         string
}

func (db *DB) UserExists(email string) (bool, error) {
	var id int
	isExists := true
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			isExists = false
		} else {
			log.Println(err)
		}
	}
	return isExists, err
}

func (db *DB) FindUserByEmail(email string) (User, error) {
	var user User
	var err error
	err = db.QueryRow("SELECT name, email, id FROM users WHERE email = ?", email).Scan(&user.Name, &user.Email, &user.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
		}
	}
	return user, err
}

func (db *DB) AddNewUser(user *User) error {
	var err error
	stmt, err := db.Prepare("INSERT INTO users(name, email, facebookId) VALUES(?, ?, ?)")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(user.Name, user.Email, user.FacebookID)
	if err != nil {
		log.Println(err)
	}
	return err
}

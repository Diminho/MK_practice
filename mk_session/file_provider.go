package mk_session

import (
	"encoding/gob"
	"log"
	"os"
	"time"
)

type FileProvider struct {
	filename       string
	expiration     int64
	expirationTime time.Time
}

func (fp *FileProvider) Init() error {
	file, err := os.OpenFile(fp.filename, os.O_CREATE, 0666)

	if err != nil {
		log.Println("Failed to open log file: ", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	return err
}

func (fp *FileProvider) Save(key, value string) error {
	file, err := os.OpenFile(fp.filename, os.O_RDWR|os.O_APPEND, 0666)

	if err != nil {
		log.Println("Failed to open log file: ", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	data := make(map[string]string)

	//Get data in case we already have this key...
	decoder := gob.NewDecoder(file)
	decoder.Decode(&data)

	// ... and re-write it
	data[key] = value

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Println(err)
	}
	return err
}

func (fp *FileProvider) Read(key string) (string, error) {
	var err error
	file, err := os.OpenFile(fp.filename, os.O_RDONLY, 0666)

	if err != nil {
		log.Println("Failed to open log file: ", err)
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	data := make(map[string]string)

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		log.Println(err)
	}

	return data[key], err

}

func (fp *FileProvider) Delete(key string) error {
	var err error

	file, err := os.OpenFile(fp.filename, os.O_RDWR, 0666)

	if err != nil {
		log.Println("Failed to open log file: ", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	data := make(map[string]string)

	//Get data in case we already have this key...
	decoder := gob.NewDecoder(file)
	decoder.Decode(&data)

	delete(data, key)

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Println(err)
	}

	return err
}

func (fp *FileProvider) EraseByExpiration() {
	err := os.Remove(fp.filename)
	if err != nil {
		log.Println("Failed to remove file: ", err)
	}
}

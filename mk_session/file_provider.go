package mk_session

import (
	"encoding/gob"
	"os"
	"time"

	"github.com/pkg/errors"
)

type FileProvider struct {
	filename       string
	expiration     int64
	expirationTime time.Time
}

func (fp *FileProvider) Init() (err error) {
	file, err := os.OpenFile(fp.filename, os.O_CREATE, 0666)

	if err != nil {
		return
	}

	defer func() {
		err = file.Close()
	}()

	return
}

func (fp *FileProvider) Save(key string, value interface{}) (err error) {
	file, err := os.OpenFile(fp.filename, os.O_RDWR|os.O_APPEND, 0666)

	if err != nil {
		return errors.Wrap(err, "cannot create session file")
	}
	defer func() {
		err = file.Close()
	}()

	data := make(map[string]interface{})

	//Get data in case we already have this key...
	decoder := gob.NewDecoder(file)
	decoder.Decode(&data)

	// ... and re-write it
	data[key] = value

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)

	return err
}

func (fp *FileProvider) Read(key string) (value interface{}, err error) {
	file, err := os.OpenFile(fp.filename, os.O_RDONLY, 0666)

	if err != nil {
		return value, errors.Wrap(err, "cannot open session file")
	}

	defer func() {
		err = file.Close()
	}()

	data := make(map[string]interface{})

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)

	if err != nil {
		return
	}

	value = data[key]

	return value, err
}

func (fp *FileProvider) Delete(key string) (err error) {

	file, err := os.OpenFile(fp.filename, os.O_RDWR, 0666)

	if err != nil {
		return errors.Wrap(err, "cannot open session file")
	}

	defer func() {
		err = file.Close()
	}()

	data := make(map[string]interface{})

	//Get data in case we already have this key...
	decoder := gob.NewDecoder(file)
	decoder.Decode(&data)

	delete(data, key)

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)

	return err
}

func (fp *FileProvider) EraseByExpiration() (err error) {
	err = os.Remove(fp.filename)
	return
}

func generateSessFilename(sessID string) string {
	return os.TempDir() + "/sess_" + sessID
}

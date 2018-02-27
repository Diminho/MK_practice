package mk_session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"time"
)

type Session interface {
	Set(key string, value interface{}) error //set session value
	Get(key string) (interface{}, error)     //get session value
	Delete(key string) error                 //delete session value
	SessionID() string                       //back current sessionID
}

type Manager struct {
	cookieName  string //private cookiename
	maxLifeTime int
}

type Provider interface {
	Save(string, interface{}) error
	Read(string) (interface{}, error)
	Delete(key string) error
	EraseByExpiration()
	Init() error
}

type Instance struct {
	sessID   string
	provider Provider
}

func NewManager(cookieName string, maxLifeTime int) *Manager {
	return &Manager{cookieName: cookieName, maxLifeTime: maxLifeTime}
}

func (mng *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (ins *Instance, err error) {
	ins = &Instance{}
	cookie, errNoCookie := r.Cookie(mng.cookieName)

	if errNoCookie == nil {
		ins.sessID = cookie.Value
		ins.SetProvider(&FileProvider{filename: "../tmp/sess_" + ins.sessID})
		err = ins.provider.Init()
		if err != nil {
			return nil, err
		}
	} else {
		ins.sessID = NewSessionID()
		cookie := http.Cookie{Name: mng.cookieName, Value: ins.sessID, Expires: (time.Now().Add(time.Duration(mng.maxLifeTime) * time.Second))}
		http.SetCookie(w, &cookie)
		ins.SetProvider(&FileProvider{filename: "../tmp/sess_" + ins.sessID})
		//delete session when time expires
		go func() {
			time.AfterFunc(time.Duration(mng.maxLifeTime)*time.Second, func() { ins.provider.EraseByExpiration() })
		}()

	}

	return
}

func (ins *Instance) SetProvider(p Provider) {
	ins.provider = p
}

func (ins *Instance) Set(key string, value interface{}) error {
	err := ins.provider.Save(key, value)

	if err != nil {
		log.Println(err)
	}

	return err
}

func (ins *Instance) Delete(key string) error {
	err := ins.provider.Delete(key)

	if err != nil {
		log.Println(err)
	}

	return err
}

func (ins *Instance) Get(key string) (value interface{}, err error) {
	value, err = ins.provider.Read(key)

	if err != nil {
		log.Println(err)
	}

	return
}

func (ins *Instance) SessionID() string {
	return ins.sessID
}

func NewSessionID() string {
	b := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

package mk_session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
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
	EraseByExpiration() error
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

	cookie, err := r.Cookie(mng.cookieName)
	if err != nil && err != http.ErrNoCookie {
		return nil, errors.Wrap(err, "cannot get cookie")
	}

	if err == http.ErrNoCookie {
		ins.sessID = NewSessionID()
		cookie := http.Cookie{Name: mng.cookieName, Value: ins.sessID, Expires: (time.Now().Add(time.Duration(mng.maxLifeTime) * time.Second))}
		http.SetCookie(w, &cookie)
		ins.SetProvider(&FileProvider{filename: generateSessFilename(ins.sessID)})
		err = ins.provider.Init()
		if err != nil {
			return nil, err
		}
		_ = time.AfterFunc(time.Duration(mng.maxLifeTime)*time.Second, func() { _ = ins.provider.EraseByExpiration() })
		return
	}

	ins.sessID = cookie.Value
	ins.SetProvider(&FileProvider{filename: generateSessFilename(ins.sessID)})

	return
}

func (ins *Instance) SetProvider(p Provider) {
	ins.provider = p
}

func (ins *Instance) Set(key string, value interface{}) error {
	return errors.Wrap(ins.provider.Save(key, value), "cannot save value")
}

func (ins *Instance) Delete(key string) error {

	return errors.Wrap(ins.provider.Delete(key), "cannot delete value")
}

func (ins *Instance) Get(key string) (value interface{}, err error) {
	value, err = ins.provider.Read(key)
	err = errors.Wrap(err, "cannot read value")
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

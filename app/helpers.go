package app

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Diminho/MK_practice/models"

	"github.com/Diminho/MK_practice/mk_session"
	"github.com/gorilla/websocket"
)

func PopulateTemplate(tmpl *template.Template, data interface{}, w http.ResponseWriter, tmplName string) {
	tmpl.ExecuteTemplate(w, tmplName, data)
}

func InStringSlice(input []string, needle string) bool {
	for _, elem := range input {
		if elem == needle {
			return true
		}
	}

	return false
}

func DeleteClient(a []*websocket.Conn, i int) {
	if i != -1 {
		copy(a[i:], a[i+1:])
		a[len(a)-1] = nil // or the zero value of T
		a = a[:len(a)-1]
	}
}

func IndexOfClient(conn *websocket.Conn, data []*websocket.Conn) int {
	for k, value := range data {
		if conn == value {
			return k
		}
	}

	return -1
}

func BuildUserIdentity(remoteAddr string) string {
	host, port, _ := net.SplitHostPort(remoteAddr)
	userID := fmt.Sprintf("%s_%s", host, port)

	return userID
}

func AuthUser(r *http.Request, s mk_session.Session, user *models.User) (err error) {
	if errSet := s.Set("email", user.Email); errSet != nil {
		err = errSet
	}

	if errSet := s.Set("isLogged", 1); errSet != nil {
		err = errSet
	}

	if errSet := s.Set("userdID", user.ID); errSet != nil {
		err = errSet
	}
	return
}

func IsLogged(r *http.Request, s mk_session.Session) (bool, error) {
	isLogged, err := s.Get("isLogged")
	if err != nil {
		return false, err
	}
	if isLogged == 1 {
		return true, nil
	}
	return false, nil
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseTemplates(dir string) (*template.Template, error) {
	var err error
	tmpl := template.New("")
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = tmpl.ParseFiles(path)

			if err != nil {
				return err
			}
		}

		return err
	})

	return tmpl, err
}

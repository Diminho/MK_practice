package app

import (
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

func PopulateTemplate(data interface{}, w http.ResponseWriter, tmplFile string) {
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		fmt.Errorf("error %v", err)
	}
	if tmpl != nil {
		tmpl.Execute(w, data)
	}

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
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
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

func IsLogged(r *http.Request) (*http.Cookie, bool) {
	var isLogged bool
	cookie, errorCookie := r.Cookie("ticket_booking")
	if errorCookie == nil {
		isLogged = true
	}
	return cookie, isLogged
}

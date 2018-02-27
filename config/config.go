package config

import "golang.org/x/oauth2"

// default DB section config
const Driver string = "mysql"
const User string = "root"
const Host string = ""
const DbName string = "ticket_booking"

const ClientID string = "179520239485516"
const ClientSecret string = "f21248b36c84b3b9c22f49ded4428a1f"
const RedirectURL string = "http://localhost:8000/facebookCallback"

var Scopes []string = []string{"public_profile", "email"}

var Endpoint oauth2.Endpoint = oauth2.Endpoint{
	AuthURL:  "https://www.facebook.com/v2.11/dialog/oauth",
	TokenURL: "https://graph.facebook.com/v2.11/oauth/access_token",
}

var FBState string = "MK_PRACTICE"
var ServerRootDir string = "./public"

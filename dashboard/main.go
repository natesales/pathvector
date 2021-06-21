package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

const (
	PeeringDbProfileUrl = "https://auth.peeringdb.com/profile/v1"
	PeeringDbAuthUrl    = "https://auth.peeringdb.com/oauth2/authorize/"
	PeeringDbTokenUrl   = "https://auth.peeringdb.com/oauth2/token/"
)

func main() {
	oauth := oauth2.Config{
		ClientID:     os.Getenv("PATHVECTOR_PDB_CLIENT_ID"),
		ClientSecret: os.Getenv("PATHVECTOR_PDB_CLIENT_SECRET"),
		RedirectURL:  "https://localhost/api/auth/redirect",
		Endpoint: oauth2.Endpoint{
			AuthURL:  PeeringDbAuthUrl,
			TokenURL: PeeringDbTokenUrl,
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := oauth.AuthCodeURL(PeeringDbProfileUrl)
		http.Redirect(w, r, u, http.StatusFound)
	})

	http.HandleFunc("/api/auth/redirect", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.Form.Get("state")
		if state != PeeringDbProfileUrl {
			w.Write([]byte("Invalid state"))
			return
		}

		token, err := oauth.Exchange(context.Background(), r.Form.Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if token == nil {
			w.Write([]byte("Authentication error"))
			return
		}
	})

	log.Println("Starting API server")
	log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil))
}

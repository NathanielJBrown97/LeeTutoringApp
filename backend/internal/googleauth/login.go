// File: backend/internal/googleauth/login.go
package googleauth

import (
	"log"
	"net/http"
	"text/template"
)

// LoginHandler serves the initial login page
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

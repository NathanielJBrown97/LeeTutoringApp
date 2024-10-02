// backend/internal/microsoftauth/login.go

package microsoftauth

import (
	"log"
	"net/http"
	"text/template"
)

func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the login template (adjust the path if necessary)
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Print("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the template and render it to the response
	if err := tmpl.Execute(w, nil); err != nil {
		log.Print("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

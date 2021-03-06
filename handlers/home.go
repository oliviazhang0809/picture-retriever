package handlers

import (
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/oliviazhang/picture-retriever/libhttp"
	"github.com/oliviazhang/picture-retriever/models"
)

func GetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	sessionStore := r.Context().Value("sessionStore").(sessions.Store)

	session, _ := sessionStore.Get(r, "picture-retriever-session")
	currentUser, ok := session.Values["user"].(*models.ImageRow)
	if !ok {
		http.Redirect(w, r, "/logout", 302)
		return
	}

	data := struct {
		CurrentUser *models.ImageRow
	}{
		currentUser,
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html.tmpl", "templates/home.html.tmpl")
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}

	tmpl.Execute(w, data)
}

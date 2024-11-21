package main

import (
	"db/internal/sqlite"
	"github.com/gofrs/uuid/v5"
	"html/template"
	"log"
	"net/http"
	"time"
)

func (app *app) SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/register.html", "/register")
}

func (app *app) SigninPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/signin.html", "/signin")
}

func (app *app) SignInHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}

	id, name, err := app.users.Authentication(
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		RenderingErrorMsg(w, "Invalid Credentials", "./assets/templates/signin.html", r)
		w.WriteHeader(400)
		return
	}

	email := r.PostForm.Get("email")
	uniqueInput := email + time.Now().Format(time.RFC3339Nano)
	sessionValue := uuid.NewV5(uuid.NamespaceURL, uniqueInput).String()

	_, err = app.users.DB.Exec(sqlite.UpdateSimiliarSessions, id)
	if err != nil {
		log.Println(err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionValue,
		Value:    sessionValue,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	_, err = app.users.DB.Exec(sqlite.InsertSession, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) StoreUserHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}
	if r.PostForm.Get("password") != r.PostForm.Get("re-password") {
		RenderingErrorMsg(w, "Passwords Don't Match", "./assets/templates/register.html", r)
		w.WriteHeader(400)
		return
	}
	err := app.users.Insert(
		r.PostForm.Get("name"),
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		if err != nil && err.Error() == "invalid email format" {
			RenderingErrorMsg(w, "Invalid Email Format", "./assets/templates/register.html", r)
			w.WriteHeader(400)
			return
		} else {
			RenderingErrorMsg(w, "Email or Username already in use", "./assets/templates/register.html", r)
			w.WriteHeader(400)
			return
		}
	}
	id, name, _ := app.users.Authentication(
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)

	uniqueInput := r.PostForm.Get("email") + time.Now().Format(time.RFC3339Nano)
	sessionValue := uuid.NewV5(uuid.NamespaceURL, uniqueInput).String()
	expiresAt := time.Now().Add(1 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionValue,
		Value:    sessionValue,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	_, err = app.users.DB.Exec(sqlite.InsertSession, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *app) EditUsernameHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		ErrorHandle(w, 500, "Error Parsing Form")
		log.Println(err)
	}
	username := r.PostForm.Get("name")

	userID, err := app.users.GetUserID(r)
	if err != nil {
		ErrorHandle(w, 500, "Error fetching userID")
		log.Println("Error fetching userID", err)
		return
	}

	_, err = app.users.DB.Exec(sqlite.ChangeUsernameQuery, username, userID)
	if err != nil {
		ErrorHandle(w, 500, "Error changing Username")
		log.Println("Error changing Username", err)
		return
	}
	_, err = app.users.DB.Exec(sqlite.ChangeUserNameInSessionsQuery, username, userID)
	if err != nil {
		ErrorHandle(w, 500, "Error changing Username")
		log.Println("Error changing Username", err)
		return
	}
	_, err = app.users.DB.Exec(sqlite.ChangeUsernameInPostsQuery, username, userID)
	if err != nil {
		ErrorHandle(w, 500, "Error changing Username in posts table")
		log.Println("Error changing Username in posts table", err)
		return
	}

	http.Redirect(w, r, "/Profile-page", http.StatusFound)

}

func RenderingErrorMsg(w http.ResponseWriter, errorMsg, path string, r *http.Request) {
	data := struct {
		ErrorMsg string
	}{
		ErrorMsg: errorMsg,
	}
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		log.Println("Error parsing template:", err)
		ErrorHandle(w, 500, "Internal Server Error")
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		ErrorHandle(w, 500, "Internal Server Error")
		return
	}

}

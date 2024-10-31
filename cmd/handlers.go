package main

import (
	// "db/internal/models"

	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
)

func render(w http.ResponseWriter, r *http.Request, t string) {
	tmp, err := template.ParseFiles(t)
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}
	tmp.Execute(w, nil)
}

func ErrorHandle(w http.ResponseWriter, statusCode int) {
	tmp, err := template.ParseFiles("./assets/templates/error.html")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Println(err)
		return
	}
	w.WriteHeader(statusCode)
	tmp.Execute(w, statusCode)
}

func (app *app) ErrorpageHandler(w http.ResponseWriter, r *http.Request){
	render(w ,r,"./assets/templates/error.html")
}

func (app *app) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.posts.AllPosts()
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}

	tmp, err := template.ParseFiles("./assets/templates/home.html")
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}

	err = tmp.Execute(w, map[string]any{"Posts": posts})
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}
}

func (app *app) ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	commentPosts, err := app.posts.PostWithComment(r)
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}
	tmp, err := template.ParseFiles("./assets/templates/post.html")
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}
	err = tmp.Execute(w, map[string]any{"info": commentPosts})
	if err != nil {
		ErrorHandle(w, 500)
		log.Println(err)
		return
	}
}

func (app *app) SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/register.html")
}
func (app *app) SigninPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/signinpage.html")
}
func (app *app) CreatPostPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/creatpostpage.html")
}

func (app *app) StoreUserHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	err = app.users.Insert(
		r.PostForm.Get("name"),
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "Email Used", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/signin", http.StatusFound)

}

func (app *app) SignInHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	id, err := app.users.Authentication(
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Println("User ID:", id)

	sessionValue := uuid.NewV5(uuid.NamespaceURL, r.PostForm.Get("email")).String()
	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionValue,
		Value:    sessionValue,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	expiresAt := time.Now().Add(1 * time.Hour)
	stmt := `INSERT OR REPLACE INTO SESSIONS (cookie_value, user_id, expires_at) VALUES (?, ?, ?)`
	_, err = app.users.DB.Exec(stmt, sessionValue, id, expiresAt)
	if err != nil {
		log.Println("Error inserting sessions:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *app) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	image, _, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer image.Close()

	timestamp := time.Now().UnixNano()
	imagePath := fmt.Sprintf("../uploads/image_%d.jpg", timestamp)
	saveImage := fmt.Sprintf("assets/uploads/image_%d.jpg", timestamp)

	place, err := os.Create(saveImage)

	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer place.Close()
	if _, err := io.Copy(place, image); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	err = app.posts.InsertPost(app.users, w, r, title, content, imagePath)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *app) SaveCommentHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	content := r.PostFormValue("content")
	postID := r.FormValue("post_id")
	err = app.posts.InsertComment(app.users, w, r, content, postID)
	if err != nil {
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)

}

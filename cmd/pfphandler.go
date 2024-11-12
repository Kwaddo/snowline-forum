package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (app *app) ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/Profile-page" {
			userData, err := app.users.AllUsersPosts(w, r)
			if err != nil {
				http.Redirect(w, r, "/signin", http.StatusFound) // redirect to signin if there's an error
				log.Println(err)
				return
			}

			tmp, err := template.ParseFiles("./assets/templates/profile.html")
			if err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}

			if err := tmp.Execute(w, map[string]any{"Users": userData}); err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}
		} else {
			ErrorHandle(w, 404, "Page not Found")
		}
	} else {
		ErrorHandle(w, 405, "Method Not Allowed")
	}
}

func (app *app) FilteringPosts(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filter := r.FormValue("action")

	var userData interface{}
	var err error

	// Determine the correct filter and fetch the data
	switch filter {
	case "like":
		userData, err = app.users.AllUserLikedPosts(w, r)
		if err != nil {
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	case "comment":
		userData, err = app.users.AllUserCommentedPosts(w, r)
		if err != nil {
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	case "dislike":
		userData, err = app.users.AllUserDisLikedPosts(w, r)
		if err != nil {
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	case "created":
		userData, err = app.users.AllUsersPosts(w, r)
		if err != nil {
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	default:
		userData, err = app.users.AllUsersPosts(w, r)
		if err != nil {
			http.Redirect(w,r,"/signin",http.StatusFound)
			return
		}
	}

	tmp, err := template.ParseFiles("./assets/templates/profile.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := tmp.Execute(w, map[string]any{"Users": userData}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

}

func (app *app) ProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	var imagePath string
	image, _, err := r.FormFile("image")
	if err != nil && err.Error() != "http: no such file" {
		ErrorHandle(w, 400, "Error retrieving the file")
		log.Println(err)
		return
	}
	if err == nil {
		defer image.Close()
		timestamp := time.Now().UnixNano()
		saveImage := fmt.Sprintf("assets/uploads/image_%d.jpg", timestamp)

		dbimage := fmt.Sprintf("../uploads/image_%d.jpg", timestamp)

		place, err := os.Create(saveImage)
		if err != nil {
			ErrorHandle(w, 500, "Unable to create file")
			log.Println(err)
			return
		}
		defer place.Close()

		if _, err := io.Copy(place, image); err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
		imagePath = dbimage
	} else {
		imagePath = ""
	}
	userID, err := app.users.GetUserID(r)
	if err != nil {
		log.Println(err)
	}
	stmt := `UPDATE USERS SET image_path = ? WHERE user_id = ?`
	_, err = app.posts.DB.Exec(stmt, imagePath, userID)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusFound)

}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (app *app) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}

	var imagePath string

	title := r.FormValue("title")
	content := r.FormValue("content")
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

	err = app.posts.InsertPost(app.users, w, r, title, content, imagePath)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *app) SaveCommentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}
	_, err := app.users.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	content := r.FormValue("content")
	postID := r.FormValue("post_id")
	const maxContentLength = 2000 // Example: 10,000 characters
	if len(content) > maxContentLength {
		ErrorHandle(w, 400, "Comment is too long")
		return
	}
	if err := app.posts.InsertComment(app.users, w, r, content, postID); err != nil {
		log.Println(err)
		ErrorHandle(w, 500, "Failed to save comment")
		return
	}
	http.Redirect(w, r, "/view-post?id="+postID, http.StatusFound)
}

func (app *app) LikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	userID, err := app.users.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = app.posts.ToggleLike(w, r, postID, userID)
	if err != nil {
		log.Println("Error toggling like:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/#post-"+postID, http.StatusFound)
}

func (app *app) DislikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	userID, err := app.users.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = app.posts.ToggleDislike(w, r, postID, userID)
	if err != nil {
		log.Println("Error toggling dislike:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/#post-"+postID, http.StatusFound)
}

func (app *app) CommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	commentID := r.FormValue("comment_id")
	userID, err := app.users.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = app.posts.ToggleCommentLike(w, r, commentID, userID)
	if err != nil {
		log.Println("Error toggling comment like:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	postID := r.FormValue("post_id")
	http.Redirect(w, r, "/view-post?id="+postID+"#comment-"+commentID, http.StatusFound)
}

func (app *app) CommentDislikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	commentID := r.FormValue("comment_id")
	userID, err := app.users.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = app.posts.ToggleCommentDislike(w, r, commentID, userID)
	if err != nil {
		log.Println("Error toggling comment like:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	postID := r.FormValue("post_id")
	http.Redirect(w, r, "/view-post?id="+postID+"#comment-"+commentID, http.StatusFound)
}

package main

import (
	"db/internal/models"
	"db/internal/sqlite"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func (app *app) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}

	image, _, err := r.FormFile("image")
	if err != nil && err.Error() != "http: no such file" {
		ErrorHandle(w, 400, "Error retrieving the file")
		log.Println(err)
		return
	}
	var imagePath string

	title := r.FormValue("title")
	content := r.FormValue("content")
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
	post_id, err := app.posts.InsertPost(app.users, w, r, title, content, imagePath)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	category_ids := r.Form["category"]
	if len(category_ids) == 0 {
		category_ids = append(category_ids, "6")
	}
	for _, category_id_str := range category_ids {

		category_id, err := strconv.Atoi(category_id_str)
		if err != nil {
			log.Println("Failed to convert category_id:", category_id_str)
			ErrorHandle(w, 400, "Invalid category_id")
			return
		}

		_, err = app.posts.DB.Exec(sqlite.InsertIntoCategory, category_id, post_id)
		if err != nil {
			log.Println("Error inserting category_id:", category_id)
			log.Println(err)
			ErrorHandle(w, 500, "Internal Server Error")
			return
		}
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
	commentID, err := app.posts.InsertComment(app.users, w, r, content, postID)
	if err != nil {
		log.Println(err)
		ErrorHandle(w, 500, "Failed to save comment")
		return
	}
	http.Redirect(w, r, "/view-post?id="+postID+"#comment-"+commentID, http.StatusFound)
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

func (app *app) FilterPosts(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	categories := r.Form["category"]
	P := []models.Post{}

	if len(categories) == 0 {
		P, err = app.posts.AllPosts()
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
	}

	var postIDs []int
	for _, cat := range categories {
		rows, err := app.users.DB.Query(sqlite.PostCategory, cat)
		if err != nil {
			ErrorHandle(w, 500, "Error querying post IDs")
			log.Println(err)
			log.Println("Error querying post IDs:", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var postID int
			if err := rows.Scan(&postID); err != nil {
				ErrorHandle(w, 500, "Error saving the file")
				log.Println(err)
				log.Println("Error scanning post ID:", err)
				return
			}

			unique := true
			for _, id := range postIDs {
				if postID == id {
					unique = false
				}
			}
			if unique {
				postIDs = append(postIDs, postID)
			}
		}
	}

	for _, ID := range postIDs {
		p := models.Post{}
		rows, err := app.users.DB.Query(sqlite.PostWithCommentQuery, ID)
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
		var postUserID string
		err = app.users.DB.QueryRow(sqlite.UserIDByPostStmt, ID).Scan(&postUserID)
		if err != nil {
			log.Println("Error fetching user_id:", err)
			return
		}

		err = app.users.DB.QueryRow(sqlite.UserProfilePicStmt, postUserID).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching profile picture:", err)
			return
		}
		for rows.Next() {
			err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
			if err != nil {
				ErrorHandle(w, 500, "Error saving the file")
				log.Println(err)
				return
			}
			P = append(P, p)
		}
	}

	sort.Slice(P, func(i, j int) bool {
		return P[i].ID > P[j].ID
	})

	if app.users.IsAuthenticated(r) {
		tmp, err := template.ParseFiles("./assets/templates/home.html")
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
		err = tmp.Execute(w, map[string]any{"Posts": P})
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
	} else {
		tmp, err := template.ParseFiles("./assets/templates/guest.html")
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
		err = tmp.Execute(w, map[string]any{"Posts": P})
		if err != nil {
			ErrorHandle(w, 500, "Error saving the file")
			log.Println(err)
			return
		}
	}
}

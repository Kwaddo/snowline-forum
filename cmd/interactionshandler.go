package main

import (
	"database/sql"
	"db/internal/models"
	"db/internal/sqlite"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

var allowedCategories = map[string]bool{
	"Sports": true,
	"Gaming": true,
	"Art":    true,
	"Music":  true,
	"Food":   true,
	"Random": true,
}

var validMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/jpg":  true,
}

func (app *app) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}
	category_ids := r.Form["category"]
	if len(category_ids) == 0 {
		category_ids = append(category_ids, "Random")
	}
	var invalidCategories []string
	for _, category := range category_ids {
		if !allowedCategories[category] {
			invalidCategories = append(invalidCategories, category)
		}
	}
	if len(invalidCategories) > 0 {
		fmt.Printf("Invalid categories: %v\n", invalidCategories)
		category_ids = []string{"Random"}
	}
	categoryIdsStr := strings.Join(category_ids, ", ")
	var imagePath string
	title := r.FormValue("title")
	content := r.FormValue("content")
	trimcontent := strings.TrimSpace(content)
	if len(trimcontent) == 0 || len(trimcontent) > 2000 {
		ErrorHandle(w, 400, "Empty Content/ Exceeded Limit")
		log.Println("Empty Content/ Exceeded Limit")
		return
	}
	const maxImageSize = 20 * 1024 * 1024
	image, header, err := r.FormFile("image")
	if err != nil && err.Error() != "http: no such file" {
		ErrorHandle(w, 400, "Error retrieving the file")
		log.Println(err)
		return
	}
	if err == nil {
		defer image.Close()
		if header.Size > maxImageSize {
			ErrorHandle(w, 400, "File size too large")
			return
		}
		contentType := header.Header.Get("Content-Type")
		if !validMimeTypes[contentType] {
			ErrorHandle(w, 400, "Invalid file type. Only JPG, JPEG, PNG, or GIF allowed")
			log.Println("Invalid file type:", contentType)
			return
		}
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
	post_id, err := app.posts.InsertPost(app.users, w, r, title, content, imagePath, categoryIdsStr)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	if len(category_ids) == 0 {
		category_ids = append(category_ids, "Random")
	}
	for _, category_id_str := range category_ids {
		_, err = app.posts.DB.Exec(sqlite.InsertIntoCategory, category_id_str, post_id)
		if err != nil {
			log.Println("Error inserting category_id:", category_id_str)
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
	trimcontent := strings.TrimSpace(content)
	if len(trimcontent) == 0 {
		ErrorHandle(w, 400, "Empty Content")
		log.Println("Empty Content")
		return
	}
	postID := r.FormValue("post_id")
	const maxContentLength = 2000 // Example: 10,000 characters
	if len(content) > maxContentLength {
		ErrorHandle(w, 400, "Comment is too long")
		return
	}
	if content == "" {
		ErrorHandle(w, 400, "Comment is empty")
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
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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

func (app *app) PostLikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	http.Redirect(w, r, "/view-post?id="+postID, http.StatusFound)
}

func (app *app) PostDislikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	http.Redirect(w, r, "/view-post?id="+postID, http.StatusFound)
}

func (app *app) ProfileLikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	http.Redirect(w, r, "/Profile-page", http.StatusFound)
}

func (app *app) ProfileDislikeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	postID := r.FormValue("post_id")
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	http.Redirect(w, r, "/Profile-page", http.StatusFound)
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
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	if postID == "" {
		log.Println("Missing post ID")
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Missing post ID")
		return
	}
	exists, err := app.posts.PostExists(postID)
	if err != nil {
		log.Println("Error checking post existence:", err)
		ErrorHandle(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if !exists {
		log.Println("Post ID does not exist:", postID)
		ErrorHandle(w, http.StatusBadRequest, "Bad Request: Post does not exist")
		return
	}
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
	var validCategories []string
	for _, category := range categories {
		if allowedCategories[category] {
			validCategories = append(validCategories, category)
		}
	}
	if len(validCategories) == 0 {
		P, err = app.posts.AllPosts()
		if err != nil {
			ErrorHandle(w, 500, "Error retrieving posts")
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

			cat := ""
			err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username, &cat)
			if err != nil {
				log.Println(err)
				if err == sql.ErrNoRows {
					continue
				}
				return
			}
			slicecat := []string{}
			slicecat = strings.Split(cat, ", ")
			for _, cat := range slicecat {
				cat = fmt.Sprintf("../images/%s.png", cat)
				p.Category = append(p.Category, cat)
			}
			err = app.posts.FetchLikesAndDislikes(&p)
			if err != nil {
				log.Println("Error fetching likes/dislikes:", err)
				return
			}

			err = app.users.DB.QueryRow(sqlite.PostCommentsCountStmt, p.ID).Scan(&p.Comments)
			if err != nil {
				log.Println("Error fetching post likes count:", err)
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

func (app *app) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error parsing form:", err)
		ErrorHandle(w, 400, "Failed to parse form")
		return
	}
	postID := r.FormValue("post_id")
	formUsername := r.FormValue("username")
	if err != nil {
		log.Println("Error fetching userID:", err)
		ErrorHandle(w, 500, "Failed to fetch user information")
		return
	}
	var postUserID int
	var postUsername string
	err = app.users.DB.QueryRow(sqlite.UserIDByPostStmt, postID).Scan(&postUserID)
	if err != nil {
		log.Println("Error fetching post's userID:", err)
		ErrorHandle(w, 500, "Failed to fetch post user ID")
		return
	}
	err = app.users.DB.QueryRow(sqlite.UserNAMEByPostStmt, postID).Scan(&postUsername)
	if err != nil {
		log.Println("Error fetching post's username:", err)
		ErrorHandle(w, 500, "Failed to fetch post username")
		return
	}
	if formUsername != postUsername {
		log.Println("Unauthorized user trying to delete post.")
		ErrorHandle(w, 403, "You are not authorized to delete this post")
		return
	}
	_, err = app.users.DB.Exec(sqlite.DeletePostQuery, postID)
	if err != nil {
		log.Println("Error deleting post:", err)
		ErrorHandle(w, 500, "Failed to delete post")
		return
	}
	_, err = app.users.DB.Exec(sqlite.DeletePostCatQuery, postID)
	if err != nil {
		log.Println("Error deleting post:", err)
		ErrorHandle(w, 500, "Failed to delete post")
		return
	}
	_, err = app.users.DB.Exec(sqlite.DeletePostlikeQuery, postID)
	if err != nil {
		log.Println("Error deleting post:", err)
		ErrorHandle(w, 500, "Failed to delete post")
		return
	}
	
	query, err := app.users.DB.Query(sqlite.CommentIDQuery, postID)
	if err != nil {
		log.Println("Error deleting post:", err)
		ErrorHandle(w, 500, "Failed to delete post")
		return
	}
	c_ids := []int{}
	for query.Next() {
		id := 0
		err = query.Scan(&id)
		if err != nil {
			ErrorHandle(w, 500, "Failed to delete post")
			return
		}
		c_ids = append(c_ids, id)
	}
	for _ , ids := range c_ids {
		_ , err =app.users.DB.Exec(sqlite.DeleteCommentlikeQuery,ids)
		if err != nil {
			ErrorHandle(w, 500, "Failed to delete post lol")
			return
		}
	}
	_, err = app.users.DB.Exec(sqlite.DeletePostcommentQuery, postID)
	if err != nil {
		log.Println("Error deleting post:", err)
		ErrorHandle(w, 500, "Failed to delete post")
		return
	}


	http.Redirect(w, r, "/Profile-page", http.StatusFound)
}

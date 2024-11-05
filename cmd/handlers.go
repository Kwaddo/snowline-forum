package main

import (
	"db/internal/sqlite"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
)

func render(w http.ResponseWriter, r *http.Request, t, urlpath string) {
	tmp, err := template.ParseFiles(t)
	if err != nil {
		ErrorHandle(w, 500, "Internal Server Error")
		log.Println(err)
		return
	}
	if r.Method == http.MethodGet {
		if r.URL.Path == urlpath {
			if err := tmp.Execute(w, nil); err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
			}
		} else {
			ErrorHandle(w, 404, "Page not Found")
		}
	} else {
		ErrorHandle(w, 405, "Method Not Allowed")
	}
}

func ErrorHandle(w http.ResponseWriter, statusCode int, message string) {
	tmp, err := template.ParseFiles("./assets/templates/error.html")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Println(err)
		return
	}
	w.WriteHeader(statusCode)
	errData := map[string]interface{}{
		"Code": statusCode,
		"Msg":  message,
	}
	if err := tmp.Execute(w, errData); err != nil {
		log.Println("Error executing error template:", err)
	}
}

func (app *app) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/" {
			posts, err := app.posts.AllPosts()
			if err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}
			if app.users.IsAuthenticated(r) {
				tmp, err := template.ParseFiles("./assets/templates/home.html")
				if err != nil {
					ErrorHandle(w, 500, "Internal Server Error")
					log.Println(err)
					return
				}

				if err := tmp.Execute(w, map[string]any{"Posts": posts}); err != nil {
					ErrorHandle(w, 500, "Internal Server Error")
					log.Println(err)
					return
				}
			} else {
				tmp, err := template.ParseFiles("./assets/templates/guest.html")
				if err != nil {
					ErrorHandle(w, 500, "Internal Server Error")
					log.Println(err)
					return
				}

				if err := tmp.Execute(w, map[string]any{"Posts": posts}); err != nil {
					ErrorHandle(w, 500, "Internal Server Error")
					log.Println(err)
					return
				}
			}

		} else {
			ErrorHandle(w, 404, "Page not Found")
		}
	} else {
		ErrorHandle(w, 405, "Method Not Allowed")
	}
}

func (app *app) ViewPostPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/view-post" {
			commentPosts, err := app.posts.PostWithComment(r)
			if err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}
			tmp, err := template.ParseFiles("./assets/templates/posts.html")
			if err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}
			if err := tmp.Execute(w, map[string]any{"info": commentPosts}); err != nil {
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

func (app *app) SignupPageHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "./assets/templates/register.html", "/register")
}

func (app *app) StoreUserHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}
	if r.PostForm.Get("password") != r.PostForm.Get("re-password") {
		ErrorHandle(w, 400, "Passwords do not match")
		return
	}

	err := app.users.Insert(
		r.PostForm.Get("name"),
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		ErrorHandle(w, 409, "Email or Username already in use ")
		return
	}
	http.Redirect(w, r, "/#login", http.StatusFound)
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
		ErrorHandle(w, 401, "Invalid credentials")
		return
	}

	email := r.PostForm.Get("email")
	uniqueInput := email + time.Now().Format(time.RFC3339Nano)
	sessionValue := uuid.NewV5(uuid.NamespaceURL, uniqueInput).String()
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

	_, err = app.users.DB.Exec(sqlite.InsertOrReplaceSession, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) SavePostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ErrorHandle(w, 400, "Failed to parse form")
		log.Println(err)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	image, _, err := r.FormFile("image")
	if err != nil {
		ErrorHandle(w, 400, "Error retrieving the file")
		log.Println(err)
		return
	}
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

	err = app.posts.InsertPost(app.users, w, r, title, content, dbimage)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "#login", http.StatusFound)
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

	content := r.FormValue("content")
	postID := r.FormValue("post_id")

	if err := app.posts.InsertComment(app.users, w, r, content, postID); err != nil {
		log.Println(err)
		ErrorHandle(w, 500, "Failed to save comment")
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *app) ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/Profile-page" {
			posts, err := app.users.AllUsersPosts(w, r)
			if err != nil {
				http.Redirect(w, r, "/#login", http.StatusFound)
				log.Println(err)
				return
			}

			tmp, err := template.ParseFiles("./assets/templates/profilepage.html")
			if err != nil {
				ErrorHandle(w, 500, "Internal Server Error")
				log.Println(err)
				return
			}

			if err := tmp.Execute(w, map[string]any{"Posts": posts}); err != nil {
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
		http.Redirect(w, r, "/#login", http.StatusFound)
		return
	}

	_, err = app.posts.DB.Exec(sqlite.InsertOrReplaceLike, postID, userID)
	if err != nil {
		http.Redirect(w, r, "/#login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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
		http.Redirect(w, r, "/#login", http.StatusFound)
		return
	}
	_, err = app.posts.DB.Exec(sqlite.InsertOrReplaceDislike, postID, userID)
	if err != nil {
		http.Redirect(w, r, "/#login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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
		log.Fatalf("Error")
	}
	stmt := `INSERT OR REPLACE INTO COMMENT_LIKES (comment_id, user_id, isliked) VALUES (?, ?, TRUE)`
	app.posts.DB.Exec(stmt, commentID, userID)
	postID := r.FormValue("post_id")
	redirectURL := fmt.Sprintf("http://localhost:8080/view-post?id=%s", postID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
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
		log.Fatalf("Error")
	}
	stmt := `INSERT OR REPLACE INTO COMMENT_LIKES (comment_id, user_id, isliked) VALUES (?, ?, FALSE)`
	app.posts.DB.Exec(stmt, commentID, userID)
	postID := r.FormValue("post_id")
	redirectURL := fmt.Sprintf("http://localhost:8080/view-post?id=%s", postID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

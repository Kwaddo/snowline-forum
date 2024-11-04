package main

import (
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
<<<<<<< HEAD

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
=======
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

>>>>>>> 1af720d85fd52d8b48d50ba7e92b3686165a9f9e
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
			tmp, err := template.ParseFiles("./assets/templates/post.html")
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

	err := app.users.Insert(
		r.PostForm.Get("name"),
		r.PostForm.Get("email"),
		r.PostForm.Get("password"),
	)
	if err != nil {
		log.Println(err)
		ErrorHandle(w, 409, "Email already in use")
		return
	}
<<<<<<< HEAD
	http.Redirect(w, r, "/signin", http.StatusFound)
=======
	http.Redirect(w, r, "/#login", http.StatusFound)
>>>>>>> 1af720d85fd52d8b48d50ba7e92b3686165a9f9e
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
	stmt := `INSERT OR REPLACE INTO SESSIONS (cookie_value, user_id, expires_at, username) VALUES (?, ?, ?, ?)`
	_, err = app.users.DB.Exec(stmt, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
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
<<<<<<< HEAD
	dbimage:= fmt.Sprintf("../uploads/image_%d.jpg", timestamp)
=======
	dbimage := fmt.Sprintf("../uploads/image_%d.jpg", timestamp)
>>>>>>> 1af720d85fd52d8b48d50ba7e92b3686165a9f9e

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
<<<<<<< HEAD
		http.Redirect(w,r,"#login",http.StatusFound)
=======
		http.Redirect(w, r, "#login", http.StatusFound)
>>>>>>> 1af720d85fd52d8b48d50ba7e92b3686165a9f9e
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
    stmt := `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, TRUE)`
    _, err = app.posts.DB.Exec(stmt, postID, userID)
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
    stmt := `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, FALSE)`
    _, err = app.posts.DB.Exec(stmt, postID, userID)
    if err != nil {
        http.Redirect(w, r, "/#login", http.StatusFound)
        return
    }
    http.Redirect(w, r, "/", http.StatusFound)
}


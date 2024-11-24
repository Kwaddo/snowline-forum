package main

import (
	"html/template"
	"log"
	"net/http"
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

	errData := map[string]interface{}{
		"Code": statusCode,
		"Msg":  message,
	}
	w.WriteHeader(statusCode)
	if err := tmp.Execute(w, errData); err != nil {
		http.Error(w, "Internal Server Error", 500)
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

func (app *app) UploadHanlder(w http.ResponseWriter, r *http.Request) {
	ErrorHandle(w, 405, "Method Not Allowed")
}

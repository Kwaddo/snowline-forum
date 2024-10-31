package main

import "net/http"

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HomepageHandler)
	mux.HandleFunc("GET /errorpage", app.ErrorpageHandler)
	mux.HandleFunc("POST /1", app.LogoutHandler)
	mux.HandleFunc("POST /2", app.SaveCommentHandler)
	// mux.HandleFunc("GET /signin", app.SigninPageHandler)
	mux.HandleFunc("POST /3", app.SignInHandler)
	mux.HandleFunc("GET /register", app.SignupPageHandler)
	mux.HandleFunc("POST /register", app.StoreUserHandler)
	// mux.HandleFunc("GET /create-post", app.CreatPostPageHandler)
	mux.HandleFunc("POST /4", app.SavePostHandler)
	mux.HandleFunc("GET /view-post", app.ViewPostHandler)
	fs := http.FileServer(http.Dir("./assets/static"))
	fs2 := http.FileServer(http.Dir("./assets/uploads"))
	fs3 := http.FileServer(http.Dir("./assets/images"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fs))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads", fs2))
	mux.Handle("GET /images/", http.StripPrefix("/images", fs3))
	return mux
}

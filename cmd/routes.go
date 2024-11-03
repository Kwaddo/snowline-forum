package main

import "net/http"

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HomepageHandler)
<<<<<<< HEAD
	mux.HandleFunc("POST /logout", app.LogoutHandler)
	mux.HandleFunc("POST /savecomment", app.SaveCommentHandler)
	mux.HandleFunc("POST /signin", app.SignInHandler)
=======
	mux.HandleFunc("GET /errorpage", app.ErrorpageHandler)
	mux.HandleFunc("POST /like", app.LikeHandler)
	mux.HandleFunc("POST /dislike", app.DislikeHandler)
	mux.HandleFunc("POST /1", app.LogoutHandler)
	mux.HandleFunc("POST /2", app.SaveCommentHandler)
	// mux.HandleFunc("GET /signin", app.SigninPageHandler)
	mux.HandleFunc("POST /3", app.SignInHandler)
>>>>>>> 82a54ecc43d67fa7a24420d5fdc03ede5b4b75a0
	mux.HandleFunc("GET /register", app.SignupPageHandler)
	mux.HandleFunc("POST /register", app.StoreUserHandler)
	mux.HandleFunc("POST /save-post", app.SavePostHandler)
	mux.HandleFunc("GET /view-post", app.ViewPostPageHandler)
	mux.HandleFunc("GET /Profile-page", app.ProfilePageHandler)

	fs := http.FileServer(http.Dir("./assets/static"))
	fs2 := http.FileServer(http.Dir("./assets/uploads"))
	fs3 := http.FileServer(http.Dir("./assets/images"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fs))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads", fs2))
	mux.Handle("GET /images/", http.StripPrefix("/images", fs3))
	return mux
}

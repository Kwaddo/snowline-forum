package main

import "net/http"

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HomepageHandler)
	mux.HandleFunc("POST /logout", app.LogoutHandler)
	mux.HandleFunc("POST /savecomment", app.SaveCommentHandler)
	mux.HandleFunc("POST /signin", app.SignInHandler)
	mux.HandleFunc("GET /signin", app.SigninPageHandler)
	mux.HandleFunc("GET /register", app.SignupPageHandler)
	mux.HandleFunc("POST /register", app.StoreUserHandler)
	mux.HandleFunc("POST /save-post", app.SavePostHandler)
	mux.HandleFunc("GET /view-post", app.ViewPostPageHandler)
	mux.HandleFunc("GET /Profile-page", app.ProfilePageHandler)
	mux.HandleFunc("POST /Profile-page/filter", app.FilteringPosts)
	mux.HandleFunc("POST /like", app.LikeHandler)
	mux.HandleFunc("POST /dislike", app.DislikeHandler)
	mux.HandleFunc("POST /profile-like", app.ProfileLikeHandler)
	mux.HandleFunc("POST /profile-dislike", app.ProfileDislikeHandler)
	mux.HandleFunc("POST /comment-like", app.CommentLikeHandler)
	mux.HandleFunc("POST /comment-dislike", app.CommentDislikeHandler)
	mux.HandleFunc("POST /profile-picture", app.ProfilePictureHandler)
	mux.HandleFunc("POST /filterposts", app.FilterPosts)
	mux.HandleFunc("POST /delete-post", app.DeletePostHandler)
	mux.HandleFunc("POST /edit-username", app.EditUsernameHandler)
	mux.HandleFunc("GET /signin/github/callback",app.handleGitHubCallback)
	mux.HandleFunc("GET /signin/google/callback",app.handleGoogleCallback)
	mux.HandleFunc("GET /signin/google/login", app.handleGoogleLogin)
	mux.HandleFunc("GET /signin/github/login", app.handleGitHubLogin)

	fs := http.FileServer(http.Dir("./assets/static"))
	fs2 := http.FileServer(http.Dir("./assets/uploads"))
	fs3 := http.FileServer(http.Dir("./assets/images"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fs))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads", fs2))
	mux.Handle("GET /images/", http.StripPrefix("/images", fs3))
	return mux
}

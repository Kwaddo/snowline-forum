package main

import (
	"db/internal/sqlite"
	"log"
	"net/http"
	"strings"
	"time"
)

func (app *app) CleanupExpiredSessions() {
	for {
		time.Sleep(time.Minute)
		app.mu.Lock()

		_, err := app.users.DB.Exec(sqlite.DeleteExpiredSessionsQuery, time.Now())
		if err != nil {
			log.Println("Error deleting expired sessions:", err)
		}

		app.mu.Unlock()
	}
}

func (app *app) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookies := r.Cookies()
	var userId string
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			row := app.posts.DB.QueryRow(sqlite.IsAuthenticateds, cookie.Value)
			row.Scan(&userId)
			if userId != "" {
				userId = cookie.Value
			} else {
				continue
			}
		}
	}

	

	if userId == "" {
		log.Println("No session cookie found")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	
	app.mu.Lock()
	defer app.mu.Unlock()

	_, err := app.users.DB.Exec(sqlite.DeleteSessionQuery, userId)
	if err != nil {
		ErrorHandle(w, 500, "Internal Server Error")
		log.Println("Error deleting session:", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + userId,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/#login", http.StatusFound)
}

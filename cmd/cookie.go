package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func (app *app) CleanupExpiredSessions() {
	for {
		time.Sleep(time.Minute)
		app.mu.Lock()

		stmt := `DELETE FROM SESSIONS WHERE expires_at < ?`
		_, err := app.users.DB.Exec(stmt, time.Now())
		if err != nil {
			log.Println(err)
			return
		}

		app.mu.Unlock()
	}
}

func (app *app) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	var sessionID string
	cookies := r.Cookies()

	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			sessionID = cookie.Value
			break
		}
	}

	if sessionID == "" {
		log.Println("No session cookie found")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	_, err := app.users.DB.Exec("DELETE FROM SESSIONS WHERE cookie_value = ?", sessionID)
	if err != nil {
		ErrorHandle(w, 500, "Internal Server Error")
		log.Println("Error deleting session:", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionID,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/signin", http.StatusFound)

}

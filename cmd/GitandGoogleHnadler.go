package main

import (
	"db/internal/sqlite"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	clientID        = "1051890298132-gdmtdqr3ub9apdoecs76skmiadq82hc7.apps.googleusercontent.com"
	clientSecret    = "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"
	redirectURI     = "http://localhost:3333/signin/google/callback"
	gitclientID     = "Ov23liNHs5SgW6S6mCcB"
	gitclientSecret = "e4d9b44188a8f1da7bf67188938a4fac06681f8d"
	gitRedirectURI  = "http://localhost:3333/signin/github/callback"
)

func (app *app) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile&state=%s",
		url.QueryEscape(clientID), url.QueryEscape(redirectURI), "random")
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (app *app) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != "random" {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}

	// Exchange the code for a token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		http.Error(w, "Failed to request token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		http.Error(w, "Access token not found", http.StatusInternalServerError)
		return
	}

	NewApp().SetGoogleUserInfo(w, r, accessToken)
}

func (app *app) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		url.QueryEscape(gitclientID),
		url.QueryEscape(gitRedirectURI),
		"random",
	)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (app *app) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != "random" {
		http.Error(w, "Invalid OAuth states", http.StatusBadRequest)
		log.Println("lol")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/signin/github/callback", http.StatusSeeOther)
	}

	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", gitclientID)
	data.Set("client_secret", gitclientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", gitRedirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Log the full response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	log.Printf("GitHub Token Response: %s", string(body)) // Log the response

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "Failed to parse token response", http.StatusInternalServerError)
		return
	}

	// Now check if "access_token" is in the response
	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		http.Error(w, "Access token not found", http.StatusInternalServerError)
		return
	}

	// Fetch GitHub user info
	NewApp().SetGitHubUserInfo(w, r, accessToken)
}

func (app *app) SetGoogleUserInfo(w http.ResponseWriter, r *http.Request, accessToken string) {
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(accessToken)
	resp, err := http.Get(userInfoURL)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	email := userInfo["email"].(string)
	username := userInfo["name"].(string)

	exists, err := app.users.CheckEmailExists(email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		err = app.users.InsertUser(email, username, "")
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}
	}

	uniqueInput := r.PostForm.Get("email") + time.Now().Format(time.RFC3339Nano)
	sessionValue := uuid.NewV5(uuid.NamespaceURL, uniqueInput).String()
	expiresAt := time.Now().Add(1 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionValue,
		Value:    sessionValue,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	id, _ := app.users.GetUserID(r)
	name, _ := app.users.GetUserName(r)
	_, err = app.users.DB.Exec(sqlite.InsertSession, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) SetGitHubUserInfo(w http.ResponseWriter, r *http.Request, accessToken string) {
	userInfoURL := "https://api.github.com/user"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		http.Error(w, "Failed to create user info request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	username, ok := userInfo["login"].(string)
	if !ok {
		http.Error(w, "Username not found", http.StatusInternalServerError)
		return
	}

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		email, err = app.fetchGitHubEmail(accessToken)
		if err != nil {
			http.Error(w, "Failed to retrieve email", http.StatusInternalServerError)
			return
		}
	}

	exists, err := app.users.CheckEmailExists(email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		err = app.users.InsertUser(email, username, "")
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}
	}

	uniqueInput := r.PostForm.Get("email") + time.Now().Format(time.RFC3339Nano)
	sessionValue := uuid.NewV5(uuid.NamespaceURL, uniqueInput).String()
	expiresAt := time.Now().Add(1 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "Forum-" + sessionValue,
		Value:    sessionValue,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	id, _ := app.users.GetUserID(r)
	name, _ := app.users.GetUserName(r)
	_, err = app.users.DB.Exec(sqlite.InsertSession, sessionValue, id, expiresAt, name)
	if err != nil {
		log.Println("Error inserting session:", err)
		ErrorHandle(w, 500, "Failed to create session")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) fetchGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	return "", fmt.Errorf("no primary verified email found")
}

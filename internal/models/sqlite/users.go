package sqlite

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type USERMODEL struct {
	DB *sql.DB
}

func (u *USERMODEL) Insert(name, email, password string) error {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Println(err)
		return err
	}

	statement := `INSERT INTO USERS (name, email, password) VALUES (?, ?, ?)`
	_, err = u.DB.Exec(statement, name, email, passwordHashed)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (u *USERMODEL) Authentication(email, password string) (int, error) {
	var id int
	var passwordHash []byte

	stmt := `SELECT user_id, password FROM USERS WHERE email = ?`
	row := u.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &passwordHash)
	if err != nil {
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if err != nil {
		return 0, err
	}
	return id, nil

}

func (u *USERMODEL) GetUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	var userID string
	cookies := r.Cookies()

	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			userID = cookie.Value
			break
		}
	}
	if userID == "" {
		log.Println("No session id found")
	}

	var id string
	stmt, err := u.DB.Prepare("SELECT user_id FROM SESSIONS WHERE cookie_value = ?")
	if err != nil {
		return "", err
	}
	row := stmt.QueryRow(userID)
	err = row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil

}

func (u *USERMODEL) GetPostID(w http.ResponseWriter, r *http.Request) (string, error) {
	var userID string
	cookies := r.Cookies()

	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			userID = cookie.Value
			break
		}
	}
	if userID == "" {
		log.Println("No session id found")
	}

	var id string
	stmt, err := u.DB.Prepare("SELECT user_id FROM SESSIONS WHERE cookie_value = ?")
	if err != nil {
		return "", err
	}
	row := stmt.QueryRow(userID)
	err = row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

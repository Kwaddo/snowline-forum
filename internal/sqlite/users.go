// user_model.go
package sqlite

import (
	"database/sql"
	"errors"
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

	_, err = u.DB.Exec(InsertUserQuery, name, email, passwordHashed)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (u *USERMODEL) Authentication(email, password string) (int, string, error) {
	var id int
	var name string
	var passwordHash []byte

	row := u.DB.QueryRow(AuthenticateUserQuery, email, email)
	err := row.Scan(&id, &passwordHash, &name)
	if err != nil {
		return 0, "", err
	}
	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if err != nil {
		return 0, "", err
	}
	return id, name, nil
}

func (u *USERMODEL) GetUserID(r *http.Request) (string, error) {
	cookies := r.Cookies()
	var cookievalue string
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			var IsValid bool
			row := u.DB.QueryRow(IsAuthenticateds, cookie.Value)
			row.Scan(&cookievalue, &IsValid)
			if cookievalue != "" && IsValid {
				cookievalue = cookie.Value
				break
			} else {
				cookievalue = ""
			}
		}
	}
	
	if cookievalue == "" {
		return "", errors.New("userId cannot be empty")
	}
	var id string
	row := u.DB.QueryRow(GetUserIDQuery, cookievalue)
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (u *USERMODEL) GetUserName(r *http.Request) (string, error) {
	cookies := r.Cookies()
	var cookievalue string
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			var IsValid bool
			row := u.DB.QueryRow(IsAuthenticateds, cookie.Value)
			row.Scan(&cookievalue, &IsValid)
			if cookievalue != "" && IsValid {
				cookievalue = cookie.Value
				break
			} else {
				cookievalue = ""
			}
		}
	}
	if cookievalue == "" {
		return "", errors.New("userId cannot be empty")
	}

	var name string
	row := u.DB.QueryRow(GetUserNameQuery, cookievalue)
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (u *USERMODEL) IsAuthenticated(r *http.Request) bool {
	cookies := r.Cookies()

	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "Forum-") {
			var value string
			var IsValid bool
			row := u.DB.QueryRow(IsAuthenticateds, cookie.Value)
			row.Scan(&value, &IsValid)
			if value != "" && IsValid {
				return true
			} else {
				continue
			}
		}
	}

	return false
}

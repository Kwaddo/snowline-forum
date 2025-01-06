// user_model.go
package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"fmt"
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
	pattern := `^[^@]+@[^@]+\.[a-zA-Z]{2,}$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	if !re.MatchString(email) {
		return fmt.Errorf("invalid email format")
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

func (u *USERMODEL) Authentication2(email, username string) (int, string, error) {
	var id int
	var name string

	row := u.DB.QueryRow(AuthenticateUserQuery2, email, username)
	err := row.Scan(&id, &name)
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

func (u *USERMODEL) CheckEmailExists(email string) (bool, error) {
	var count int
	row := u.DB.QueryRow(CheckEmailExistsQuery, email)
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Println("Error checking email existence:", err)
		return false, err
	}

	return count > 0, nil
}

func (u *USERMODEL) CheckNameExists(username string) (bool, error) {
	var count int
	row := u.DB.QueryRow(CheckUsernameExistsQuery, username)
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Println("Error checking email existence:", err)
		return false, err
	}

	return count > 0, nil
}

func (u *USERMODEL) InsertUser(name, email, password string) error {
	_, err := u.DB.Exec(InsertUserQuery, name, email, password)
	if err != nil {
		log.Println("Error inserting user:", err)
		return err
	}
	return nil
}

func (u *USERMODEL) GetUserRole(w http.ResponseWriter, r *http.Request) (string, error) {
	var role string
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching user_id:", err)
		return "", err
	}
	row := u.DB.QueryRow(CheckUserRoleByUserIDQuery, userID)
	err2 := row.Scan(&role)
	if err2 != nil {
		log.Println("Error checking user role:", err)
		return "", err
	}
	return role, nil
}

func (u *USERMODEL) PromoteUserToAdmin(w http.ResponseWriter, r *http.Request) error {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching user_id:", err)
		return err
	}
	_, err = u.DB.Exec(ChangeRoleToAdminQuery, userID)
	if err != nil {
		log.Println("Error promoting user to admin:", err)
		return err
	}
	return nil
}

func (u *USERMODEL) PromoteUserToModerator(w http.ResponseWriter, r *http.Request) error {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching user_id:", err)
		return err
	}
	_, err = u.DB.Exec(ChangeRoleToModeratorQuery, userID)
	if err != nil {
		log.Println("Error promoting user to admin:", err)
		return err
	}
	return nil
}

func (u *USERMODEL) DemoteUserToNormal(w http.ResponseWriter, r *http.Request) error {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching user_id:", err)
		return err
	}
	_, err = u.DB.Exec(ChangeRoleToUserQuery, userID)
	if err != nil {
		log.Println("Error demoting user from admin:", err)
		return err
	}
	return nil
}

func (u *USERMODEL) GetUserRoleByID(userID string) (string, error) {
	var role string
	row := u.DB.QueryRow(CheckUserRoleByUserIDQuery, userID)
	err := row.Scan(&role)
	if err != nil {
		log.Println("Error checking user role:", err)
		return "", err
	}
	return role, nil
}

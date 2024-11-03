package sqlite

import (
	"database/sql"
	"db/internal/models"
	"log"
	"net/http"
)

type POSTMODEL struct {
	DB *sql.DB
}

func (m *POSTMODEL) InsertPost(userModel *USERMODEL, w http.ResponseWriter, r *http.Request, title, content, image_path string) error {
	userID, err := userModel.GetUserID(w, r)
	if err != nil {
		log.Println(err)
		return err
	}
	stmt := `INSERT INTO POSTS (title, content, image_path, user_id, created_at) VALUES (?, ?, ?, ?, datetime('now'))`
	_, err = m.DB.Exec(stmt, title, content, image_path, userID)
	return err
}

func (m *POSTMODEL) InsertComment(userModel *USERMODEL, w http.ResponseWriter, r *http.Request, content, post_id string) error {
	userID, err := userModel.GetUserID(w, r)
	if err != nil {
		log.Println(err)
		return err
	}
	stmt := `INSERT INTO COMMENTS (post_id, user_id, content, created_at) VALUES (?, ?, ?, datetime('now'))`
	_, err = m.DB.Exec(stmt, post_id, userID, content)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func (m *POSTMODEL) AllPosts() ([]models.Post, error) {
	stmt := `SELECT post_id, title, content, image_path, created_at FROM POSTS ORDER BY post_id DESC`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	posts := []models.Post{}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content,&p.ImagePath, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		likesStmt := `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = TRUE`
			err = m.DB.QueryRow(likesStmt, p.ID).Scan(&p.Likes)
			if err != nil {
				log.Println("Error fetching likes count:", err)
			}
		dislikesStmt := `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = FALSE`
		err = m.DB.QueryRow(dislikesStmt, p.ID).Scan(&p.Dislikes)
		if err != nil {
			log.Println("Error fetching dislikes count:", err)
		}
		posts = append(posts, p)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return posts, nil
}

func (u *USERMODEL ) AllUsersPosts(w http.ResponseWriter, r *http.Request) ([]models.Post, error) {
	stmt := `SELECT post_id, title, content, image_path, created_at FROM POSTS WHERE user_id = ? ORDER BY post_id DESC`
	userID, err := u.GetUserID(w,r)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	rows, err := u.DB.Query(stmt, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	posts := []models.Post{}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content,&p.ImagePath, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return posts, nil
}

func (m *POSTMODEL) PostWithComment(r *http.Request) (models.PostandComment, error) {
	postID := r.URL.Query().Get("id")

	stmt := `SELECT post_id, title, content, image_path, created_at FROM POSTS WHERE post_id = ?`
	row := m.DB.QueryRow(stmt, postID)

	p := models.Post{}
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt)
	if err != nil {
		log.Println(err)
		return models.PostandComment{}, err
	}

	stmt2 := `SELECT comment_id, post_id, content, created_at FROM COMMENTS WHERE post_id = ?`
	rows, err := m.DB.Query(stmt2, postID)
	if err != nil {
		log.Println(err)
		return models.PostandComment{}, err
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		c := models.Comment{}
		err := rows.Scan(&c.ID, &c.PostID, &c.Content, &c.CreatedAt)
		if err != nil {
			log.Println(err)
			return models.PostandComment{}, err
		}
		comments = append(comments, c)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return models.PostandComment{}, err
	}

	// Construct the PostandComment struct
	commentPost := models.PostandComment{
		Posts:   p,
		Comment: comments,
	}

	return commentPost, nil
}

// your original file, e.g., post_model.go
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
	userID, err := userModel.GetUserID(r)
	if err != nil {
		log.Println(err)
		return err
	}
	userName, err := userModel.GetUserName(r)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = m.DB.Exec(InsertPostQuery, title, content, image_path, userID, userName)
	return err
}

func (m *POSTMODEL) InsertComment(userModel *USERMODEL, w http.ResponseWriter, r *http.Request, content, post_id string) error {
	userID, err := userModel.GetUserID(r)
	if err != nil {
		log.Println(err)
		return err
	}
	userName, err := userModel.GetUserName(r)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = m.DB.Exec(InsertCommentQuery, post_id, userID, content, userName)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *POSTMODEL) AllPosts() ([]models.Post, error) {
	rows, err := m.DB.Query(AllPostsQuery)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			return nil, err
		}
		err = m.fetchLikesAndDislikes(&p)
		if err != nil {
			log.Println("Error fetching likes/dislikes:", err)
		}
		commentsStmt := `SELECT COUNT(*) from COMMENTS WHERE post_id = ?`
		err = m.DB.QueryRow(commentsStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
		}

		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return posts, nil
}

func (m *POSTMODEL) fetchLikesAndDislikes(p *models.Post) error {
	err := m.DB.QueryRow(PostLikesCountQuery, p.ID).Scan(&p.Likes)
	if err != nil {
		return err
	}
	err = m.DB.QueryRow(PostDislikesCountQuery, p.ID).Scan(&p.Dislikes)
	return err
}

func (u *USERMODEL) AllUsersPosts(w http.ResponseWriter, r *http.Request) ([]models.Post, error) {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rows, err := u.DB.Query(AllUsersPostsQuery, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return posts, nil
}

func (m *POSTMODEL) PostWithComment(r *http.Request) (models.PostandComment, error) {
    postID := r.URL.Query().Get("id")


    stmt := `SELECT post_id, title, content, image_path, created_at, UserName FROM POSTS WHERE post_id = ?`
    row := m.DB.QueryRow(stmt, postID)

    p := models.Post{}
    err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
    if err != nil {
        log.Println(err)
        return models.PostandComment{}, err
    }

   
    likesStmt := `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = TRUE`
    err = m.DB.QueryRow(likesStmt, p.ID).Scan(&p.Likes)
    if err != nil {
        log.Println("Error fetching post likes count:", err)
    }

 
    dislikesStmt := `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = FALSE`
    err = m.DB.QueryRow(dislikesStmt, p.ID).Scan(&p.Dislikes)
    if err != nil {
        log.Println("Error fetching post dislikes count:", err)
    }
	commentsStmt := `SELECT COUNT(*) from COMMENTS WHERE post_id = ?`
		err = m.DB.QueryRow(commentsStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
		}

    stmt2 := `SELECT comment_id, post_id, content, created_at, username FROM COMMENTS WHERE post_id = ?`
    rows, err := m.DB.Query(stmt2, postID)
    if err != nil {
        log.Println(err)
        return models.PostandComment{}, err
    }
    defer rows.Close()

    comments := []models.Comment{}
    for rows.Next() {
        c := models.Comment{}
        err := rows.Scan(&c.ID, &c.PostID, &c.Content, &c.CreatedAt, &c.Username)
        if err != nil {
            log.Println(err)
            return models.PostandComment{}, err
        }

        commentLikesStmt := `SELECT COUNT(*) FROM COMMENT_LIKES WHERE comment_id = ? AND isliked = TRUE`
        err = m.DB.QueryRow(commentLikesStmt, c.ID).Scan(&c.Likes)
        if err != nil {
            log.Println("Error fetching comment likes count:", err)
        }


        commentDislikesStmt := `SELECT COUNT(*) FROM COMMENT_LIKES WHERE comment_id = ? AND isliked = FALSE`
        err = m.DB.QueryRow(commentDislikesStmt, c.ID).Scan(&c.Dislikes)
        if err != nil {
            log.Println("Error fetching comment dislikes count:", err)
        }

        comments = append(comments, c)
    }

    err = rows.Err()
    if err != nil {
        log.Println(err)
        return models.PostandComment{}, err
    }


    commentPost := models.PostandComment{
        Posts:    p,
        Comments: comments,
    }

    return commentPost, nil
}


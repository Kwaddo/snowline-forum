package sqlite

import (
	"database/sql"
	"db/internal/models"
	"fmt"
	"log"
	"net/http"
	"time"
)

type POSTMODEL struct {
	DB *sql.DB
}

func (m *POSTMODEL) InsertPost(userModel *USERMODEL, w http.ResponseWriter, r *http.Request, title, content, image_path string) (int64, error) {
	userID, err := userModel.GetUserID(r)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	var userName string
	err = m.DB.QueryRow(UserNamefromUserID, userID).Scan(&userName)
	if err != nil {
		log.Println("Error fetching username:", err)
		return 0, err
	}

	post_id, err := m.DB.Exec(InsertPostQuery, title, content, image_path, userID, userName, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Println(err)
		return 0, err
	}

	postID, err := post_id.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return postID, err
}

func (m *POSTMODEL) InsertComment(userModel *USERMODEL, w http.ResponseWriter, r *http.Request, content, post_id string) (string, error) {
	userID, err := userModel.GetUserID(r)
	if err != nil {
		log.Println(err)
		return "", err
	}
	userName, err := userModel.GetUserName(r)
	if err != nil {
		log.Println(err)
		return "", err
	}
	result, err := m.DB.Exec(InsertCommentQuery, post_id, userID, content, userName, time.Now().Format("2006-01-02 03:04:05"))
	if err != nil {
		log.Println(err)
		return "", err
	}
	commentID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return "", err
	}
	return fmt.Sprintf("%d", commentID), nil
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
			return nil, err
		}

		err = m.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
			return nil, err
		}
		userID := m.DB.QueryRow(UserIDByPostStmt, p.ID)
		id := ""
		err = userID.Scan(&id)
		if err != nil {
			log.Println("Error fetching user_id")
			return nil, err
		}
		err = m.DB.QueryRow(UserProfilePicStmt, id).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching image path", err)
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

func (m *POSTMODEL) fetchLikesAndDislikes(p *models.Post) error {
	err := m.DB.QueryRow(PostLikesCountQuery, p.ID).Scan(&p.Likes)
	if err != nil {
		return err
	}
	err = m.DB.QueryRow(PostDislikesCountQuery, p.ID).Scan(&p.Dislikes)
	return err
}

func (u *USERMODEL) fetchLikesAndDislikes(p *models.Post) error {
	err := u.DB.QueryRow(PostLikesCountQuery, p.ID).Scan(&p.Likes)
	if err != nil {
		return err
	}
	err = u.DB.QueryRow(PostDislikesCountQuery, p.ID).Scan(&p.Dislikes)
	return err
}

func (u *USERMODEL) AllUsersPosts(w http.ResponseWriter, r *http.Request) (models.PostandMainUsername, error) {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println(err)
		return models.PostandMainUsername{}, err
	}

	rows, err := u.DB.Query(AllUsersPostsQuery, userID)
	if err != nil {
		log.Println(err)
		return models.PostandMainUsername{}, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			return models.PostandMainUsername{}, err
		}

		err = u.DB.QueryRow(UserNAMEByPostStmt, p.ID).Scan(&p.Username)
		if err != nil {
			log.Println("Error fetching user_id")
			return models.PostandMainUsername{}, err
		}
		err = u.fetchLikesAndDislikes(&p)
		if err != nil {
			log.Println("Error fetching likes/dislikes:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
			return models.PostandMainUsername{}, err
		}

		err = u.DB.QueryRow(UserProfilePicStmt, userID).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching profile image", err)
			return models.PostandMainUsername{}, err
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return models.PostandMainUsername{}, err
	}

	username, err := u.GetUserName(r)
	if err != nil {
		log.Println("Error getting username:", err)
		return models.PostandMainUsername{}, err
	}

	var path string
	uID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching userID", err)
		return models.PostandMainUsername{}, err
	}
	err = u.DB.QueryRow(UserProfilePicStmt, uID).Scan(&path)
	if err != nil {
		log.Println("Error fetching image path", err)
		return models.PostandMainUsername{}, err
	}

	imgPath := "../uploads/delete-button.png"

	result := models.PostandMainUsername{
		Posts:     posts,
		Username:  username,
		ImagePath: path,
		Delete:    imgPath,
	}

	return result, nil
}

func (u *USERMODEL) AllUserLikedPosts(w http.ResponseWriter, r *http.Request) (models.PostandMainUsername, error) {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		return models.PostandMainUsername{}, err
	}
	username, err := u.GetUserName(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		return models.PostandMainUsername{}, err
	}
	rows, err := u.DB.Query(AllUserLikedPostsQuery, userID)
	if err != nil {
		log.Println("Error querying post IDs:", err)
		return models.PostandMainUsername{}, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			log.Println("Error scanning post ID:", err)
			return models.PostandMainUsername{}, err
		}
		postIDs = append(postIDs, postID)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error during row iteration:", err)
		return models.PostandMainUsername{}, err
	}

	var posts []models.Post
	for _, postID := range postIDs {
		row := u.DB.QueryRow(PostWithCommentQuery, postID)
		p := models.Post{}
		err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			log.Println("Error scanning post:", err)
			return models.PostandMainUsername{}, err
		}

		var postUserID string
		err = u.DB.QueryRow(UserIDByPostStmt, p.ID).Scan(&postUserID)
		if err != nil {
			log.Println("Error fetching user_id:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.fetchLikesAndDislikes(&p)
		if err != nil {
			log.Println("Error fetching likes/dislikes:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(UserProfilePicStmt, postUserID).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching profile picture:", err)
			return models.PostandMainUsername{}, err
		}

		posts = append(posts, p)
	}
	var path string
	uID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching userID", err)
		return models.PostandMainUsername{}, err
	}
	err = u.DB.QueryRow(UserProfilePicStmt, uID).Scan(&path)
	if err != nil {
		log.Println("Error fetching image path", err)
		return models.PostandMainUsername{}, err
	}
	result := models.PostandMainUsername{
		Posts:    posts,
		Username: username,
		ImagePath: path,
	}

	return result, nil
}

func (u *USERMODEL) AllUserDisLikedPosts(w http.ResponseWriter, r *http.Request) (models.PostandMainUsername, error) {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		return models.PostandMainUsername{}, err
	}
	username, err := u.GetUserName(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		return models.PostandMainUsername{}, err
	}
	rows, err := u.DB.Query(AllUserDisLikedPostsQuery, userID)
	if err != nil {
		log.Println("Error querying post IDs:", err)
		return models.PostandMainUsername{}, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			log.Println("Error scanning post ID:", err)
			return models.PostandMainUsername{}, err
		}
		postIDs = append(postIDs, postID)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error during row iteration:", err)
		return models.PostandMainUsername{}, err
	}

	var posts []models.Post
	for _, postID := range postIDs {
		row := u.DB.QueryRow(PostWithCommentQuery, postID)
		p := models.Post{}
		err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			log.Println("Error scanning post:", err)
			return models.PostandMainUsername{}, err
		}

		var postUserID string
		err = u.DB.QueryRow(UserIDByPostStmt, p.ID).Scan(&postUserID)
		if err != nil {
			log.Println("Error fetching user_id:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.fetchLikesAndDislikes(&p)
		if err != nil {
			log.Println("Error fetching likes/dislikes:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(UserProfilePicStmt, postUserID).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching profile picture:", err)
			return models.PostandMainUsername{}, err
		}

		posts = append(posts, p)
	}
	var path string
	uID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching userID", err)
		return models.PostandMainUsername{}, err
	}
	err = u.DB.QueryRow(UserProfilePicStmt, uID).Scan(&path)
	if err != nil {
		log.Println("Error fetching image path", err)
		return models.PostandMainUsername{}, err
	}
	result := models.PostandMainUsername{
		Posts:    posts,
		Username: username,
		ImagePath: path,
	}

	return result, nil
}

func (u *USERMODEL) AllUserCommentedPosts(w http.ResponseWriter, r *http.Request) (models.PostandMainUsername, error) {
	userID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error getting user ID:", err)
		return models.PostandMainUsername{}, err
	}
	username, err := u.GetUserName(r)
	if err != nil {
		log.Println("Error getting user name:", err)
		return models.PostandMainUsername{}, err
	}
	rows, err := u.DB.Query(AllUserCommentedPostsQuery, userID)
	if err != nil {
		log.Println("Error querying post IDs:", err)
		return models.PostandMainUsername{}, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			log.Println("Error scanning post ID:", err)
			return models.PostandMainUsername{}, err
		}
		postIDs = append(postIDs, postID)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error during row iteration:", err)
		return models.PostandMainUsername{}, err
	}

	var posts []models.Post
	for _, postID := range postIDs {
		row := u.DB.QueryRow(PostWithCommentQuery, postID)
		p := models.Post{}
		err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			log.Println("Error scanning post:", err)
			return models.PostandMainUsername{}, err
		}

		var postUserID string
		err = u.DB.QueryRow(UserIDByPostStmt, p.ID).Scan(&postUserID)
		if err != nil {
			log.Println("Error fetching user ID:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.fetchLikesAndDislikes(&p)
		if err != nil {
			log.Println("Error fetching likes/dislikes:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
		if err != nil {
			log.Println("Error fetching post likes count:", err)
			return models.PostandMainUsername{}, err
		}
		err = u.DB.QueryRow(UserProfilePicStmt, postUserID).Scan(&p.ProfilePic)
		if err != nil {
			log.Println("Error fetching profile picture:", err)
			return models.PostandMainUsername{}, err
		}

		posts = append(posts, p)
	}
	var path string
	uID, err := u.GetUserID(r)
	if err != nil {
		log.Println("Error fetching userID", err)
		return models.PostandMainUsername{}, err
	}
	err = u.DB.QueryRow(UserProfilePicStmt, uID).Scan(&path)
	if err != nil {
		log.Println("Error fetching image path", err)
		return models.PostandMainUsername{}, err
	}
	result := models.PostandMainUsername{
		Posts:    posts,
		Username: username,
		ImagePath: path,
	}

	return result, nil
}

func (m *POSTMODEL) PostWithComment(r *http.Request) (models.PostandComment, error) {
	postID := r.URL.Query().Get("id")

	p := models.Post{}
	row := m.DB.QueryRow(PostWithCommentQuery, postID)
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt, &p.Username)
	if err != nil {
		log.Println(err)
		return models.PostandComment{}, err
	}

	err = m.DB.QueryRow(PostLikesCountQuery, p.ID).Scan(&p.Likes)
	if err != nil {
		log.Println("Error fetching post likes count:", err)
		return models.PostandComment{}, err
	}

	err = m.DB.QueryRow(PostDislikesCountQuery, p.ID).Scan(&p.Dislikes)
	if err != nil {
		log.Println("Error fetching post dislikes count:", err)
		return models.PostandComment{}, err
	}

	err = m.DB.QueryRow(PostCommentsCountStmt, p.ID).Scan(&p.Comments)
	if err != nil {
		log.Println("Error fetching post comments count:", err)
		return models.PostandComment{}, err
	}

	var userID string
	err = m.DB.QueryRow(UserIDByPostStmt, p.ID).Scan(&userID)
	if err != nil {
		log.Println("Error fetching user_id:", err)
		return models.PostandComment{}, err
	}

	err = m.DB.QueryRow(UserProfilePicStmt, userID).Scan(&p.ProfilePic)
	if err != nil {
		log.Println("Error fetching image path:", err)
		return models.PostandComment{}, err
	}

	rows, err := m.DB.Query(CommentsForPostQuery, postID)
	if err != nil {
		log.Println(err)
		return models.PostandComment{}, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		c := models.Comment{}
		err := rows.Scan(&c.ID, &c.PostID, &c.Content, &c.CreatedAt, &c.Username)
		if err != nil {
			log.Println(err)
			return models.PostandComment{}, err
		}

		err = m.DB.QueryRow(CommentLikesCountStmt, c.ID).Scan(&c.Likes)
		if err != nil {
			log.Println("Error fetching comment likes count:", err)
			return models.PostandComment{}, err
		}

		err = m.DB.QueryRow(CommentDislikesCountStmt, c.ID).Scan(&c.Dislikes)
		if err != nil {
			log.Println("Error fetching comment dislikes count:", err)
			return models.PostandComment{}, err
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

func (m *POSTMODEL) ToggleLike(w http.ResponseWriter, r *http.Request, postID string, userID string) error {
	var isLiked *bool
	err := m.DB.QueryRow(PostIsLikedQuery, postID, userID).Scan(&isLiked)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Println("Error checking like status:", err)
		return err
	}
	if isLiked != nil && *isLiked {
		_, err = m.DB.Exec(RemoveIsLikedQuery, postID, userID)
		if err != nil {
			log.Println("Error removing like:", err)
			return err
		}
	} else {
		_, err = m.DB.Exec(InsertOrReplaceLike, postID, userID)
		if err != nil {
			log.Println("Error adding like:", err)
			return err
		}
	}
	return nil
}

func (m *POSTMODEL) ToggleDislike(w http.ResponseWriter, r *http.Request, postID string, userID string) error {
	var isLiked *bool
	err := m.DB.QueryRow(PostIsLikedQuery, postID, userID).Scan(&isLiked)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Println("Error checking like status:", err)
		return err
	}
	if isLiked != nil && !*isLiked {
		_, err = m.DB.Exec(RemoveIsLikedQuery, postID, userID)
		if err != nil {
			log.Println("Error removing like:", err)
			return err
		}
	} else {
		_, err = m.DB.Exec(InsertOrReplaceDislike, postID, userID)
		if err != nil {
			log.Println("Error adding like:", err)
			return err
		}
	}
	return nil
}

func (m *POSTMODEL) ToggleCommentLike(w http.ResponseWriter, r *http.Request, commentID string, userID string) error {
	var isLiked *bool
	err := m.DB.QueryRow(CommentIsLikedQuery, commentID, userID).Scan(&isLiked)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking like status:", err)
		return err
	}

	if isLiked != nil && *isLiked {
		_, err = m.DB.Exec(RemoveCommentIsLikedQuery, commentID, userID)
		if err != nil {
			log.Println("Error removing comment like:", err)
			return err
		}
	} else {
		_, err = m.DB.Exec(InsertOrReplaceLikeComment, commentID, userID)
		if err != nil {
			log.Println("Error adding comment like:", err)
			return err
		}
	}
	return nil
}

func (m *POSTMODEL) ToggleCommentDislike(w http.ResponseWriter, r *http.Request, commentID string, userID string) error {
	var isLiked *bool
	err := m.DB.QueryRow(CommentIsLikedQuery, commentID, userID).Scan(&isLiked)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking like status:", err)
		return err
	}
	if isLiked != nil && !*isLiked {
		_, err = m.DB.Exec(RemoveCommentIsLikedQuery, commentID, userID)
		if err != nil {
			log.Println("Error removing comment like:", err)
			return err
		}
	} else {
		_, err = m.DB.Exec(InsertOrReplaceDislikeComment, commentID, userID)
		if err != nil {
			log.Println("Error adding comment like:", err)
			return err
		}
	}
	return nil
}

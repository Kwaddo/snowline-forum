package sqlite

// Insert statements to save in DB
const (
	InsertPostQuery               = `INSERT INTO POSTS (title, content, image_path, user_id, UserName, created_at, categories) VALUES (?, ?, ?, ?, ?, ?, ?)`
	InsertCommentQuery            = `INSERT INTO COMMENTS (post_id, user_id, content, username, created_at) VALUES (?, ?, ?, ?, ?)`
	InsertUserQuery               = `INSERT INTO USERS (name, email, password) VALUES (?, ?, ?)`
	CheckEmailExistsQuery         = `SELECT COUNT(*) FROM users WHERE email = ?`
	CheckUsernameExistsQuery      = `SELECT COUNT(*) FROM users WHERE name = ?`
	InsertUserQueryNoP            = `INSERT INTO users (name, email) VALUES (?, ?)`
	InsertSession                 = `INSERT INTO SESSIONS (cookie_value, user_id, expires_at, username, isvalid) VALUES (?, ?, ?, ?, true);`
	InsertOrReplaceLike           = `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, TRUE);`
	InsertOrReplaceDislike        = `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, FALSE);`
	InsertOrReplaceLikeComment    = `INSERT OR REPLACE INTO COMMENT_LIKES (comment_id, user_id, isliked) VALUES (?, ?, TRUE)`
	InsertOrReplaceDislikeComment = `INSERT OR REPLACE INTO COMMENT_LIKES (comment_id, user_id, isliked) VALUES (?, ?, FALSE)`
	InsertIntoCategory            = `INSERT INTO POST_CATEGORIES (category_id, post_id) VALUES (?, ?)`
)

// Deactivate session statements from DB
const (
	UpdateExpiredSessionsQuery = `UPDATE SESSIONS SET isvalid = false WHERE expires_at < ?`
	UpdateSessionQuery         = `UPDATE SESSIONS SET isvalid = false WHERE cookie_value = ?`
	UpdateSimiliarSessions     = `UPDATE SESSIONS SET isvalid = false WHERE user_id = ?`
)

// Select statements
const (
	AllPostsQuery              = `SELECT post_id, title, content, image_path, created_at, UserName, categories FROM POSTS ORDER BY post_id DESC`
	AllUsersPostsQuery         = `SELECT post_id, title, content, image_path, created_at, UserName, categories FROM POSTS WHERE user_id = ? ORDER BY post_id DESC`
	PostWithCommentQuery       = `SELECT post_id, title, content, image_path, created_at, UserName, categories FROM POSTS WHERE post_id = ? ORDER BY post_id DESC`
	CommentsForPostQuery       = `SELECT comment_id, post_id, content, created_at, username FROM COMMENTS WHERE post_id = ?`
	PostLikesCountQuery        = `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = TRUE`
	PostDislikesCountQuery     = `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = FALSE`
	IsAuthenticateds           = `SELECT cookie_value, isvalid FROM SESSIONS WHERE cookie_value = ?`
	AllUserCommentedPostsQuery = `SELECT post_id FROM COMMENTS WHERE user_id = ?`
	AllUserLikedPostsQuery     = `SELECT post_id FROM POST_LIKES WHERE user_id = ? AND isliked = true`
	AllUserDisLikedPostsQuery  = `SELECT post_id FROM POST_LIKES WHERE user_id = ? AND isliked = false`
	PostCommentsCountStmt      = `SELECT COUNT(*) FROM COMMENTS WHERE post_id = ?`
	UserIDByPostStmt           = `SELECT user_id FROM POSTS WHERE post_id = ?`
	UserNAMEByPostStmt         = `SELECT username FROM POSTS WHERE post_id = ?`
	UserProfilePicStmt         = `SELECT image_path FROM USERS WHERE user_id = ?`
	CommentLikesCountStmt      = `SELECT COUNT(*) FROM COMMENT_LIKES WHERE comment_id = ? AND isliked = TRUE`
	CommentDislikesCountStmt   = `SELECT COUNT(*) FROM COMMENT_LIKES WHERE comment_id = ? AND isliked = FALSE`
	PostIsLikedQuery           = `SELECT isliked FROM POST_LIKES WHERE post_id = ? AND user_id = ?`
	CommentIsLikedQuery        = `SELECT isliked FROM COMMENT_LIKES WHERE comment_id = ? AND user_id = ?`
	PostCategory               = `SELECT post_id FROM post_categories WHERE category_id = ? ORDER BY post_id DESC`
	UserNamefromUserID         = `SELECT name FROM USERS WHERE user_id = ?`
	PostExistsQuery            = `SELECT COUNT(*) FROM POSTS WHERE post_id = ?`
)

// Update statements
const (
	RemoveIsLikedQuery            = `UPDATE POST_LIKES SET isliked = NULL WHERE post_id = ? AND user_id = ?`
	RemoveCommentIsLikedQuery     = `UPDATE COMMENT_LIKES SET isliked = NULL WHERE comment_id = ? AND user_id = ?`
	ChangeUsernameQuery           = `UPDATE USERS SET name = ? WHERE user_id = ?`
	ChangeUserNameInSessionsQuery = `UPDATE SESSIONS SET username = ? WHERE user_id = ?`
	ChangeUsernameInPostsQuery    = `UPDATE POSTS SET username = ? WHERE user_id = ?`
	
)

// Delete statements
const (
	DeletePostQuery = `DELETE FROM POSTS WHERE POST_ID = ?`
	DeletePostCatQuery = `DELETE FROM post_categories WHERE POST_ID = ?`
	DeletePostlikeQuery = `DELETE FROM Post_likes WHERE POST_ID = ?`
	DeleteCommentlikeQuery = `DELETE FROM Comment_likes WHERE COMMENT_ID = ?`
	DeletePostcommentQuery = `DELETE FROM Comments WHERE POST_ID = ?`
	CommentIDQuery = `SELECT comment_id  FROM comments WHERE POST_ID = ?`
)

// Select ---> Authentication and User Retrieval
const (
	AuthenticateUserQuery  = `SELECT user_id, password, name FROM USERS WHERE email = ? OR name = ?`
	AuthenticateUserQuery2 = `SELECT user_id, name FROM USERS WHERE email = ? OR name = ?`
	GetUserIDQuery         = `SELECT user_id FROM SESSIONS WHERE cookie_value = ?`
	GetUserNameQuery       = `SELECT username FROM SESSIONS WHERE cookie_value = ?`
)

package sqlite

// Insert statements to save in DB
const (
	InsertPostQuery        = `INSERT INTO POSTS (title, content, image_path, user_id, UserName, created_at) VALUES (?, ?, ?, ?, ?, datetime('now'))`
	InsertCommentQuery     = `INSERT INTO COMMENTS (post_id, user_id, content, username, created_at) VALUES (?, ?, ?, ?, datetime('now'))`
	InsertUserQuery        = `INSERT INTO USERS (name, email, password) VALUES (?, ?, ?)`
	InsertOrReplaceSession  = `INSERT OR REPLACE INTO SESSIONS (cookie_value, user_id, expires_at, username) VALUES (?, ?, ?, ?);`
	InsertOrReplaceLike    = `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, TRUE);`
	InsertOrReplaceDislike  = `INSERT OR REPLACE INTO POST_LIKES (post_id, user_id, isliked) VALUES (?, ?, FALSE);`
)

// Delete statements from DB
const (
	DeleteExpiredSessionsQuery = `DELETE FROM SESSIONS WHERE expires_at < ?`
	DeleteSessionQuery          = `DELETE FROM SESSIONS WHERE cookie_value = ?`
)

// Select statements 
const (
	AllPostsQuery          = `SELECT post_id, title, content, image_path, created_at, UserName FROM POSTS ORDER BY post_id DESC`
	AllUsersPostsQuery     = `SELECT post_id, title, content, image_path, created_at, UserName FROM POSTS WHERE user_id = ? ORDER BY post_id DESC`
	PostWithCommentQuery    = `SELECT post_id, title, content, image_path, created_at, UserName FROM POSTS WHERE post_id = ?`
	CommentsForPostQuery    = `SELECT comment_id, post_id, content, created_at, username FROM COMMENTS WHERE post_id = ?`
	PostLikesCountQuery     = `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = TRUE`
	PostDislikesCountQuery   = `SELECT COUNT(*) FROM POST_LIKES WHERE post_id = ? AND isliked = FALSE`
	IsAuthenticateds = `SELECT cookie_value FROM SESSIONS WHERE cookie_value = ?`
)

// Select ---> Authentication and User Retrieval
const (
	AuthenticateUserQuery = `SELECT user_id, password, name FROM USERS WHERE email = ? OR name = ?`
	GetUserIDQuery        = `SELECT user_id FROM SESSIONS WHERE cookie_value = ?`
	GetUserNameQuery      = `SELECT username FROM SESSIONS WHERE cookie_value = ?`
)

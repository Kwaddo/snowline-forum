package models

type Post struct {
	Username   string
	ID         int
	Title      string
	Content    string
	ImagePath  string
	ProfilePic string
	CreatedAt  string
	Likes      string
	Dislikes   string
	Comments   string
}

type Comment struct {
	Username  string
	ID        int
	PostID    int
	Content   string
	CreatedAt string
	Likes     string
	Dislikes  string
	Comments  string
}

type PostandComment struct {
	Posts    Post
	Comments []Comment
}

type PostandMainUsername struct {
	Posts     []Post
	Username  string
	ImagePath string
}

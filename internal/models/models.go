package models

import "time"

type Post struct {
	Username  string
	ID        int
	Title     string
	Content   string
	ImagePath string
	CreatedAt time.Time
	Likes     string
	Dislikes  string
}

type Comment struct {
	Username  string
	ID        int
	PostID    int
	Content   string
	CreatedAt time.Time
	Likes     string
	Dislikes  string
}

type PostandComment struct {
	Posts    Post
	Comments []Comment
}

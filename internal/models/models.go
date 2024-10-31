package models

import "time"

type Post struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	CreatedAt time.Time
}

type Comment struct {
	ID        int
	PostID    int
	Content   string
	CreatedAt time.Time
}

type PostandComment struct {
	Posts   Post
	Comment []Comment
}

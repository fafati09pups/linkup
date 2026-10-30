package models

import "time"

type Msg struct {
	ID       int64     `json:"id"`
	ChatID   int       `json:"chat_id"`
	SenderID int       `json:"sender_id"`
	Username string    `json:"username"`
	Content  string    `json:"content"`
	At       time.Time `json:"created_at"`
}

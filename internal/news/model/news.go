package model

import "time"

// News represents news
type News struct {
	ID      int       `json:"id" gorm:"primary_key"`
	Author  string    `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created" gorm:"default:CURRENT_TIMESTAMP"`
}

type NewsMapping struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
}

type NewsRequest struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

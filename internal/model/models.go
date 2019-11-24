package model

import "time"

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// News represents news
type News struct {
	ID      int       `json:"id" gorm:"primary_key"`
	Author  string    `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created" gorm:"default:CURRENT_TIMESTAMP"`
}

// NewsResponse represents list of news response
type NewsResponse struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
	Data  []News `json:"data"`
}

type NewsMapping struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
}

type NewsRequest struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

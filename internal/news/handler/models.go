package handler

import (
	"github.com/rogaps/newsapi/internal/news/model"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewsResponse represents list of news response
type NewsResponse struct {
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int           `json:"total"`
	Data  []*model.News `json:"data"`
}

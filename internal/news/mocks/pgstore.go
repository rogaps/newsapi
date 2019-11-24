package mocks

import (
	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/stretchr/testify/mock"
)

type PGStore struct {
	mock.Mock
}

func (s *PGStore) FindByID(id int) *model.News {
	args := s.Called(id)

	var news *model.News
	if args.Get(0) != nil {
		news = args.Get(0).(*model.News)
	}
	return news
}

func (s *PGStore) Create(news *model.News) {
}

func (s *PGStore) Count() (count int) {
	args := s.Called()
	return args.Int(0)
}

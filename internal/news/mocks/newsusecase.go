package mocks

import (
	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/stretchr/testify/mock"
)

type NewsUsecase struct {
	mock.Mock
}

func (u *NewsUsecase) GetNews(page, limit int) ([]*model.News, error) {
	args := u.Called(page, limit)
	news := args.Get(0).([]*model.News)
	return news, args.Error(0)
}

func (u *NewsUsecase) CreateNews(news *model.News) error {
	args := u.Called(news)
	return args.Error(0)
}

func (u *NewsUsecase) Count() (count int) {
	args := u.Called()
	return args.Int(0)
}

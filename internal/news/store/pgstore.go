package store

import "github.com/jinzhu/gorm"

import "github.com/rogaps/newsapi/internal/news/model"

type PGStore interface {
	FindByID(int) *model.News
	Create(*model.News)
	Count() int
}

// PGStore represents postgres store
type pgStore struct {
	db *gorm.DB
}

func NewPGStore(db *gorm.DB) PGStore {
	return &pgStore{db}
}

func (s *pgStore) FindByID(id int) *model.News {
	news := model.News{ID: id}
	s.db.Find(&news)
	return &news
}

func (s *pgStore) Create(news *model.News) {
	s.db.Create(news)
	return
}

func (s *pgStore) Count() (count int) {
	s.db.Model(model.News{}).Count(&count)
	return
}

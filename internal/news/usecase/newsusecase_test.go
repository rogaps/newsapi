package usecase

import (
	"testing"
	"time"

	"github.com/rogaps/newsapi/internal/news/mocks"
	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/rogaps/newsapi/internal/news/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetNews(t *testing.T) {
	pgStore := new(mocks.PGStore)
	esStore := new(mocks.ESStore)
	hits := []*store.Hit{
		&store.Hit{
			ID: 1,
		},
		&store.Hit{
			ID: 2,
		},
	}
	mockNews := []*model.News{
		&model.News{
			ID:      1,
			Author:  "John Doe",
			Body:    "Lorem Ipsum",
			Created: time.Date(2019, 11, 24, 10, 10, 10, 10, time.UTC),
		},
		&model.News{
			ID:      2,
			Author:  "John Doe",
			Body:    "Lorem Ipsum",
			Created: time.Date(2019, 11, 24, 10, 10, 10, 10, time.UTC),
		},
	}

	searchResults := &store.SearchResults{
		Total: 2,
		Hits:  hits,
	}

	esStore.On("Search", mock.Anything).Return(searchResults, nil)
	pgStore.On("FindByID", 1).Return(mockNews[0])
	pgStore.On("FindByID", 2).Return(mockNews[1])

	uc := NewNewsUsecase(esStore, pgStore)
	news, err := uc.GetNews(1, 10)
	assert.Nil(t, err)
	assert.Equal(t, news, mockNews)
}

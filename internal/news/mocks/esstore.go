package mocks

import (
	"github.com/rogaps/newsapi/internal/news/store"
	"github.com/stretchr/testify/mock"
)

type ESStore struct {
	mock.Mock
}

func (s *ESStore) Search(query map[string]interface{}) (*store.SearchResults, error) {
	args := s.Called(query)
	var results *store.SearchResults
	var err error
	if args.Get(0) != nil {
		results = args.Get(0).(*store.SearchResults)
	}
	err = args.Error(1)
	return results, err
}

func (s *ESStore) Create(docID string, doc interface{}) error {
	args := s.Called(docID, doc)
	return args.Error(0)
}

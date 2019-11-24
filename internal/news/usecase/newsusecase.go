package usecase

import (
	"strconv"
	"sync"

	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/rogaps/newsapi/internal/news/store"
	log "github.com/sirupsen/logrus"
)

type Usecase interface {
	GetNews(int, int) ([]*model.News, error)
	CreateNews(*model.News) error
	Count() int
}

type NewsUsecase struct {
	esStore store.ESStore
	pgStore store.PGStore
}

func NewNewsUsecase(esStore store.ESStore, pgStore store.PGStore) Usecase {
	return &NewsUsecase{esStore, pgStore}
}

func (u *NewsUsecase) GetNews(page, limit int) (news []*model.News, err error) {
	var wg sync.WaitGroup
	query := map[string]interface{}{
		"from": (page - 1) * limit,
		"size": limit,
		"sort": map[string]interface{}{
			"created": map[string]string{
				"order": "desc",
			},
		},
	}
	res, err := u.esStore.Search(query)
	if err != nil {
		return news, err
	}

	news = make([]*model.News, len(res.Hits))

	wg.Add(len(news))
	for i := range res.Hits {
		go func(i int) {
			news[i] = u.pgStore.FindByID(res.Hits[i].ID)
			wg.Done()
		}(i)
	}
	wg.Wait()

	return news, nil
}

func (u *NewsUsecase) CreateNews(news *model.News) error {
	log.Infof("Store news author: %s, body: %s to db.\n", news.Author, news.Body)
	u.pgStore.Create(news)
	doc := model.NewsMapping{
		ID:      news.ID,
		Created: news.Created,
	}
	log.Infof("Store news ID: %s, created: %s to ES.\n", news.ID, news.Created)
	docID := strconv.Itoa(doc.ID)
	err := u.esStore.Create(docID, doc)
	return err
}

func (u *NewsUsecase) Count() (count int) {
	return u.pgStore.Count()
}

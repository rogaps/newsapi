package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/go-redis/cache/v7"
	"github.com/jinzhu/gorm"
	"github.com/rogaps/newsapi/internal/model"
	"github.com/rogaps/newsapi/pkg/elasticsearch"
	"github.com/rogaps/newsapi/pkg/messaging"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// News represents news handler
type NewsHandler struct {
	db         *gorm.DB
	amqpClient *messaging.AMQPClient
	redisStore *cache.Codec
	esClient   *elasticsearch.ESClient
}

var (
	PageCachePrefix = "news:cache:"
)

func CreateKey(u string) string {
	return urlEscape(PageCachePrefix, u)
}
func urlEscape(prefix string, u string) string {
	key := url.QueryEscape(u)
	if len(key) > 200 {
		h := sha1.New()
		io.WriteString(h, u)
		key = string(h.Sum(nil))
	}
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(":")
	buffer.WriteString(key)
	return buffer.String()
}

func NewNewsHandler(db *gorm.DB, amqpClient *messaging.AMQPClient, redisStore *cache.Codec, esClient *elasticsearch.ESClient) *NewsHandler {
	return &NewsHandler{db: db, amqpClient: amqpClient, redisStore: redisStore, esClient: esClient}
}

func (h *NewsHandler) ListNews(w http.ResponseWriter, r *http.Request) {
	var newsResponse model.NewsResponse
	url := r.URL
	key := CreateKey(url.RequestURI())
	w.Header().Set("Content-Type", "application/json")
	if 1 != 1 {
		if h.redisStore.Exists(key) {
			err := h.redisStore.Get(key, &newsResponse)
			if err == nil {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&newsResponse)
				return
			}
			log.Errorln(err)
		}
	}

	params := url.Query()
	page, err := strconv.Atoi(params.Get("page"))
	if page <= 0 || err != nil {
		page = 1
	}
	limit, err := strconv.Atoi(params.Get("limit"))
	if limit <= 0 || err != nil {
		limit = 10
	}
	query := map[string]interface{}{
		"from": (page - 1) * limit,
		"size": limit,
		"sort": map[string]interface{}{
			"created": map[string]string{
				"order": "desc",
			},
		},
	}
	res, err := h.esClient.Search(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&model.ErrorResponse{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	news := make([]model.News, len(res.Hits))
	for i, hit := range res.Hits {
		news[i].ID = hit.ID
	}
	var wg sync.WaitGroup
	wg.Add(len(news))
	for i := range news {
		log.Infoln(news[i].ID)
		go func(i int) {
			h.db.Find(&news[i])
			wg.Done()
		}(i)
	}

	wg.Wait()

	w.WriteHeader(http.StatusOK)
	newsResponse.Page = page
	newsResponse.Limit = limit
	var total int
	h.db.Model(&model.News{}).Count(&total)
	newsResponse.Total = total
	newsResponse.Data = news
	h.redisStore.Set(&cache.Item{
		Key:    key,
		Object: newsResponse,
	})
	json.NewEncoder(w).Encode(&newsResponse)
}

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	var news model.NewsRequest
	json.NewDecoder(r.Body).Decode(&news)
	payload, err := json.Marshal(news)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&model.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}
	err = h.amqpClient.PublishOnQueue(payload, "news")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&model.ErrorResponse{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
}

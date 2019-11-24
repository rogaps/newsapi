package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-redis/cache/v7"
	"github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"

	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/rogaps/newsapi/internal/news/usecase"
	"github.com/rogaps/newsapi/pkg/messaging"
)

// NewsHandler represents news handler
type NewsHandler struct {
	amqpClient  *messaging.AMQPClient
	newsUsecase usecase.Usecase
	redisStore  *cache.Codec
	redisClient *redis.Client
}

func NewNewsHandler(amqpClient *messaging.AMQPClient, newsUsecase usecase.Usecase, redisStore *cache.Codec, redisClient *redis.Client) *NewsHandler {
	return &NewsHandler{
		amqpClient:  amqpClient,
		newsUsecase: newsUsecase,
		redisStore:  redisStore,
		redisClient: redisClient,
	}
}

func (h *NewsHandler) ListNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := r.URL.Query()
	page, err := strconv.Atoi(params.Get("page"))
	if page <= 0 || err != nil {
		page = 1
	}
	limit, err := strconv.Atoi(params.Get("limit"))
	if limit <= 0 || err != nil {
		limit = 10
	}

	news, err := h.newsUsecase.GetNews(page, limit)

	w.WriteHeader(http.StatusOK)
	newsResponse := &NewsResponse{
		Page:  page,
		Limit: limit,
		Total: h.newsUsecase.Count(),
		Data:  news,
	}
	json.NewEncoder(w).Encode(newsResponse)
}

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	var news model.NewsRequest
	json.NewDecoder(r.Body).Decode(&news)
	payload, err := json.Marshal(news)
	if err != nil {
		RespondError(http.StatusBadRequest, err, w, r)
		return
	}
	err = h.amqpClient.PublishOnQueue(payload, "news")
	if err != nil {
		RespondError(http.StatusInternalServerError, err, w, r)
		return
	}
	iter := h.redisClient.Scan(0, PageCachePrefix+"*", 0).Iterator()
	for iter.Next() {
		err := h.redisClient.Del(iter.Val()).Err()
		if err != nil {
			log.Errorf("failed to delete cache: %s\n", err)
		}
	}
	if err := iter.Err(); err != nil {
		log.Errorf("failed to delete cache: %s\n", err)
	}
	w.WriteHeader(http.StatusCreated)
}

func RespondError(code int, err error, w http.ResponseWriter, r *http.Request) {
	log.Errorln(err)
	errResponse := ErrorResponse{Code: code, Message: err.Error()}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errResponse)
	return
}

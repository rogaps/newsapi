package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-redis/cache/v7"
	log "github.com/sirupsen/logrus"
)

var (
	PageCachePrefix = "cache:news"
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

type CacheHandler struct {
	redisStore *cache.Codec
}

func NewCacheHandler(redisStore *cache.Codec) *CacheHandler {
	return &CacheHandler{redisStore}
}

func (h *CacheHandler) Handle(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newsResponse NewsResponse
		url := r.URL
		key := CreateKey(url.RequestURI())

		if h.redisStore.Exists(key) {
			err := h.redisStore.Get(key, &newsResponse)
			if err == nil {
				log.Infoln("Serving from cache...")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&newsResponse)
				return
			}
			log.Errorln(err)
		}
		writer := newCachedWriter(h.redisStore, time.Hour, w, key)
		next(writer, r)
	}
}

type cachedWriter struct {
	writer http.ResponseWriter
	store  *cache.Codec
	expire time.Duration
	key    string
}

func newCachedWriter(store *cache.Codec, expire time.Duration, writer http.ResponseWriter, key string) *cachedWriter {
	return &cachedWriter{writer, store, expire, key}
}

func (w *cachedWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *cachedWriter) WriteHeader(code int) {
	w.writer.WriteHeader(code)
}

func (w *cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.writer.Write(data)
	if err == nil {
		w.store.Set(&cache.Item{
			Key:        w.key,
			Object:     data,
			Expiration: w.expire,
		})
	}
	return ret, err
}

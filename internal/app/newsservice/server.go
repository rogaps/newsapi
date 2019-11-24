package newsservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/cache/v7"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rogaps/newsapi/internal/app"
	"github.com/rogaps/newsapi/internal/news/handler"
	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/rogaps/newsapi/internal/news/store"
	"github.com/rogaps/newsapi/internal/news/usecase"
	"github.com/rogaps/newsapi/pkg/messaging"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
)

// Server represents newsservice server
type Server struct {
	config       *app.Config
	newsHandler  *handler.NewsHandler
	cacheHandler *handler.CacheHandler
	db           *gorm.DB
}

// NewServer creates new server
func NewServer(config *app.Config, newsHandler *handler.NewsHandler, cacheHandler *handler.CacheHandler, db *gorm.DB) *Server {
	return &Server{
		config:       config,
		newsHandler:  newsHandler,
		cacheHandler: cacheHandler,
		db:           db,
	}
}

func (s *Server) Handler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/news", s.cacheHandler.Handle(s.newsHandler.ListNews)).Methods("GET")
	router.HandleFunc("/news", s.newsHandler.CreateNews).Methods("POST")
	return handlers.CORS()(router)
}

func (s *Server) Run() {
	s.db.AutoMigrate(&model.News{})
	log.Infoln("Starting server...")
	http.ListenAndServe(s.config.Server.Address, s.Handler())
}

// BuildContainer builds container of DI
func BuildContainer(configFile string) *dig.Container {
	container := dig.New()
	container.Provide(app.NewConfig(configFile))
	container.Provide(ConnectDatabase)
	container.Provide(ConnectBroker)
	container.Provide(ConnectRedisCache)
	container.Provide(ConstructRedisClient)
	container.Provide(ConnectElasticsearch)
	container.Provide(store.NewPGStore)
	container.Provide(usecase.NewNewsUsecase)
	container.Provide(handler.NewNewsHandler)
	container.Provide(handler.NewCacheHandler)
	container.Provide(NewServer)
	return container
}

func ConnectElasticsearch(config *app.Config) (store.ESStore, error) {
	return store.NewESSTore("news", config.Elasticsearch.ConnectionString, "", "")
}

func ConnectBroker(config *app.Config) (*messaging.AMQPClient, error) {
	return messaging.Connect(config.AMQP.ConnectionString)
}

func ConnectDatabase(config *app.Config) (*gorm.DB, error) {
	return gorm.Open("postgres", config.Postgres.ConnectionString)
}

func ConnectRedisCache(config *app.Config) *cache.Codec {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{"redis": config.Redis.Address},
	})
	return &cache.Codec{
		Redis: ring,
		Marshal: func(v interface{}) ([]byte, error) {
			if b, ok := v.([]byte); ok {
				return b, nil
			}
			return nil, fmt.Errorf("value is not json")
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return json.Unmarshal(b, v)
		},
	}
}

func ConstructRedisClient(config *app.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: config.Redis.Address,
	})
}

package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"github.com/go-redis/cache/v7"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rogaps/newsapi/internal/handler"
	"github.com/rogaps/newsapi/internal/model"
	"github.com/rogaps/newsapi/internal/newsapi"
	"github.com/rogaps/newsapi/pkg/elasticsearch"
	"github.com/rogaps/newsapi/pkg/messaging"
	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"
)

func main() {
	configFile := flag.String("c", "", "json configuration file")
	flag.Parse()

	container := dig.New()
	container.Provide(newsapi.NewConfig(*configFile))
	container.Provide(ConnectDatabase)
	container.Provide(ConnectBroker)
	container.Provide(ConnectRedis)
	container.Provide(ConnectElasticsearch)
	container.Provide(handler.NewNewsHandler)
	container.Provide(NewServer)
	err := container.Invoke(func(server *Server) {
		server.Run()
	})
	if err != nil {
		log.Errorln(err)
	}
}

func ConnectElasticsearch(config *newsapi.Config) (*elasticsearch.ESClient, error) {
	return elasticsearch.Connect("news", config.Elasticsearch.ConnectionString, "", "")
}

func ConnectBroker(config *newsapi.Config) (*messaging.AMQPClient, error) {
	return messaging.Connect(config.AMQP.ConnectionString)
}

func ConnectDatabase(config *newsapi.Config) (*gorm.DB, error) {
	return gorm.Open("postgres", config.Postgres.ConnectionString)
}

func ConnectRedis(config *newsapi.Config) *cache.Codec {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{"redis": ":6379"},
	})
	return &cache.Codec{
		Redis: ring,
		Marshal: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return json.Unmarshal(b, v)
		},
	}
}

type Server struct {
	config      *newsapi.Config
	newsHandler *handler.NewsHandler
	db          *gorm.DB
}

func NewServer(config *newsapi.Config, newsHandler *handler.NewsHandler, db *gorm.DB) *Server {
	return &Server{
		config:      config,
		newsHandler: newsHandler,
		db:          db,
	}
}

func (s *Server) Handler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/news", s.newsHandler.ListNews).Methods("GET")
	router.HandleFunc("/news", s.newsHandler.CreateNews).Methods("POST")
	return handlers.CORS()(router)
}

func (s *Server) Run() {
	s.db.AutoMigrate(&model.News{})
	log.Infoln("Starting server...")
	http.ListenAndServe(s.config.Server.Address, s.Handler())
}

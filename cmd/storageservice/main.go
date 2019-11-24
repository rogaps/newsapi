package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"go.uber.org/dig"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rogaps/newsapi/internal/model"
	"github.com/rogaps/newsapi/internal/newsapi"
	"github.com/rogaps/newsapi/pkg/elasticsearch"
	"github.com/rogaps/newsapi/pkg/messaging"
	"github.com/streadway/amqp"
)

func main() {
	configFile := flag.String("c", "", "json configuration file")
	flag.Parse()

	container := dig.New()
	container.Provide(newsapi.NewConfig(*configFile))
	container.Provide(ConnectDatabase)
	container.Provide(ConnectBroker)
	container.Provide(ConnectElasticsearch)
	container.Provide(NewStoreService)
	container.Provide(NewServer)

	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	if err := container.Invoke(func(server *Server) {
		server.Run()
	}); err != nil {
		log.Errorln(err)
	}

	<-done
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

type Server struct {
	amqpClient   *messaging.AMQPClient
	storeService *StoreService
}

func NewServer(amqpClient *messaging.AMQPClient, storeService *StoreService) *Server {
	return &Server{amqpClient, storeService}
}

func (s *Server) Run() {
	log.Infoln("Starting server...")
	s.amqpClient.SubscribeToQueue("news", "", s.storeService.Consume)
}

type StoreService struct {
	db       *gorm.DB
	esClient *elasticsearch.ESClient
}

func NewStoreService(db *gorm.DB, esClient *elasticsearch.ESClient) *StoreService {
	return &StoreService{db, esClient}
}

func (s *StoreService) Consume(d amqp.Delivery) {
	var newsReq model.NewsRequest

	json.Unmarshal(d.Body, &newsReq)

	news := model.News{
		Author: newsReq.Author,
		Body:   newsReq.Body,
	}
	s.db.Create(&news)
	doc := model.NewsMapping{
		ID:      news.ID,
		Created: news.Created,
	}
	docID := strconv.Itoa(doc.ID)
	err := s.esClient.Create(docID, doc)
	if err != nil {
		log.Errorln(err)
	}
}

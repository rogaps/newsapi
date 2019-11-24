package storageservice

import (
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

type Server struct {
	amqpClient     *messaging.AMQPClient
	messageHandler handler.MessageHandler
	db             *gorm.DB
}

func NewServer(amqpClient *messaging.AMQPClient, messageHandler handler.MessageHandler, db *gorm.DB) *Server {
	return &Server{amqpClient, messageHandler, db}
}

func (s *Server) Run() {
	log.Infoln("Starting server...")
	s.db.AutoMigrate(&model.News{})
	s.amqpClient.SubscribeToQueue("news", "", s.messageHandler.Consume)
}

// BuildContainer builds container of DI
func BuildContainer(configFile string) *dig.Container {
	container := dig.New()
	container.Provide(app.NewConfig(configFile))
	container.Provide(ConnectDatabase)
	container.Provide(ConnectBroker)
	container.Provide(ConnectElasticsearch)
	container.Provide(store.NewPGStore)
	container.Provide(usecase.NewNewsUsecase)
	container.Provide(handler.NewMessageHandler)
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

package handler

import (
	"encoding/json"

	"github.com/rogaps/newsapi/internal/news/model"
	"github.com/rogaps/newsapi/internal/news/usecase"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type MessageHandler interface {
	Consume(d amqp.Delivery)
}
type AMQPHandler struct {
	newsUsecase usecase.Usecase
}

func NewMessageHandler(newsUsecase usecase.Usecase) MessageHandler {
	return &AMQPHandler{newsUsecase}
}

func (h *AMQPHandler) Consume(d amqp.Delivery) {
	var newsReq model.NewsRequest

	json.Unmarshal(d.Body, &newsReq)

	news := &model.News{
		Author: newsReq.Author,
		Body:   newsReq.Body,
	}
	log.Infof("Consuming message: %s\n", d.Body)
	err := h.newsUsecase.CreateNews(news)
	if err != nil {
		log.Errorln(err)
	}
}

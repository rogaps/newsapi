package main

import (
	"flag"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rogaps/newsapi/internal/app/newsservice"
	log "github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("c", "", "json configuration file")
	flag.Parse()

	container := newsservice.BuildContainer(*configFile)
	err := container.Invoke(func(server *newsservice.Server) {
		server.Run()
	})
	if err != nil {
		log.Errorln(err)
	}
}

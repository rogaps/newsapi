package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rogaps/newsapi/internal/app/storageservice"
)

func main() {
	configFile := flag.String("c", "", "json configuration file")
	flag.Parse()

	container := storageservice.BuildContainer(*configFile)

	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	if err := container.Invoke(func(server *storageservice.Server) {
		server.Run()
	}); err != nil {
		panic(err)
	}

	<-done
}

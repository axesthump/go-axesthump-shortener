package main

import (
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func handleShutdown(signalHandler chan os.Signal, done chan bool, conf *config.AppConfig) {
	<-signalHandler
	err := conf.Repo.Close()
	if err != nil {
		panic(err)
	}
	done <- true
}

func main() {
	conf, err := config.CreateAppConfig()
	signalHandler := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		panic(err)
	}
	appHandler := handlers.NewAppHandler(conf.BaseURL+"/", conf.Repo)
	go handleShutdown(signalHandler, done, conf)
	go func() {
		log.Fatal(http.ListenAndServe(conf.ServerAddr, appHandler.Router))
	}()
	<-done
	os.Exit(0)
}

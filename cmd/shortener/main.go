package main

import (
	_ "github.com/jackc/pgx/v5"
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/handlers"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func handleShutdown(signalHandler chan os.Signal, done chan bool, conf *config.AppConfig) {
	<-signalHandler
	err := conf.Repo.Close()
	if err != nil {
		panic(err)
	}
	if conf.Conn != nil {
		err = conf.Conn.Close(conf.DBContext)
		if err != nil {
			panic(err)
		}
	}
	conf.DeleteService.Close()
	conf.UserIDGenerator.Cancel()
	done <- true
}

func main() {
	conf, err := config.NewAppConfig()
	signalHandler := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		panic(err)
	}
	appHandler := handlers.NewAppHandler(conf)
	go handleShutdown(signalHandler, done, conf)
	go func() {
		log.Printf("Start listen server at %s\n", conf.ServerAddr)
		log.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
		if conf.IsHTTPS {
			manager := &autocert.Manager{
				Cache:  autocert.DirCache("cache-dir"),
				Prompt: autocert.AcceptTOS,
			}
			server := &http.Server{
				Addr:      conf.ServerAddr,
				Handler:   appHandler.Router,
				TLSConfig: manager.TLSConfig(),
			}
			log.Fatal(server.ListenAndServeTLS("", ""))
		} else {
			log.Fatal(http.ListenAndServe(conf.ServerAddr, appHandler.Router))
		}
	}()
	<-done
}

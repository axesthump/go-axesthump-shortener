package main

import (
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/handlers"
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"net/http"
	"os"
)

func main() {
	os.Setenv("FILE_STORAGE_PATH", "./temp")
	conf, err := config.CreateAppConfig()
	defer func(Repo repository.Repository) {
		err := Repo.Close()
		if err != nil {
			panic(err)
		}
	}(conf.Repo)
	if err != nil {
		panic(err)
	}
	appHandler := handlers.NewAppHandler(conf.BaseURL+"/", conf.Repo)
	log.Fatal(http.ListenAndServe(conf.ServerAddr, appHandler.Router))
}

package main

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"inventory/internal/config"
	"inventory/internal/handler"
	"inventory/internal/repository"
	"inventory/internal/service"
)

func main() {
	// Загрузка конфигурации один раз при старте
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	repo := repository.NewRepository(rdb, cfg)
	serv := service.NewService(repo, cfg)
	hand := handler.NewHandler(serv, cfg)

	r := mux.NewRouter()
	// Dynamically set routes based on the loaded configuration
	for _, entity := range cfg.DatabaseStructure {
		updateEndpoint := "/" + entity.UpdateEndpoint
		getEndpoint := "/" + entity.GetEndpoint + "/{guid}"

		r.HandleFunc(updateEndpoint, hand.UpdateEntityHandler).Methods("POST")
		r.HandleFunc(getEndpoint, hand.GetEntityHandler).Methods("GET")
	}

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"inventory/internal/config"
	"inventory/internal/handler"
	"inventory/internal/repository"
	"inventory/internal/service"
)

func main() {

	cfg := config.GlobalConfig
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	repo := repository.NewRepository(rdb, cfg)
	serv := service.NewService(repo, cfg)
	hand := handler.NewHandler(serv, cfg)

	r := mux.NewRouter()
	for _, entity := range cfg.DatabaseStructure {
		updateEndpoint := "/" + entity.UpdateEndpoint
		getEndpoint := "/" + entity.GetEndpoint + "/{guid}"

		r.HandleFunc(updateEndpoint, hand.UpdateEntityHandler).Methods("POST")
		r.HandleFunc(getEndpoint, hand.GetEntityHandler).Methods("GET")
	}

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

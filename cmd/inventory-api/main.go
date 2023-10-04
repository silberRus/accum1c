package main

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"inventory/internal/config" // импортировать пакет конфигурации
	"inventory/internal/handler"
	"inventory/internal/repository"
	"inventory/internal/service"
)

func main() {
	// Загрузить конфигурационный файл
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// Используйте значения из конфигурации для настройки приложения
	// Например, используйте адрес Redis из конфигурации
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr}) // Предположим, что у вас есть поле RedisAddr в структуре Config
	repo := repository.NewRepository(rdb, cfg.DatabaseStructure)
	serv := service.NewService(repo, cfg)
	hand := handler.NewHandler(serv)

	r := mux.NewRouter()
	r.HandleFunc("/inventory/update", hand.UpdateInventory).Methods("POST")
	r.HandleFunc("/inventory/{product_id}", hand.GetInventory).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

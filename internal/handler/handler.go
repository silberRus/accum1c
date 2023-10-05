package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"inventory/internal/config"
	"net/http"

	"github.com/gorilla/mux"
	"inventory/internal/service"
)

type Handler struct {
	service *service.Service
	config  *config.Config
}

func NewHandler(service *service.Service, config *config.Config) *Handler {
	return &Handler{service: service, config: config}
}

func (h *Handler) UpdateEntityHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	entityName := ""

	for _, entityConfig := range h.config.DatabaseStructure {
		if "/"+entityConfig.UpdateEndpoint == path {
			entityName = entityConfig.Entity
			break
		}
	}

	if entityName == "" {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		return
	}

	entityConfig := h.config.GetEntityConfig(entityName)

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.UpdateEntity(entityName, request, entityConfig)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			http.Error(w, "Failed redis is nil", http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetEntityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityName := vars["entity"]

	// Validate entity
	entityConfig := h.config.GetEntityConfig(entityName)
	if entityConfig == nil {
		http.Error(w, "Invalid entity", http.StatusBadRequest)
		return
	}

	entityGUID := vars["guid"]
	entityData, err := h.service.GetEntity(entityName, entityGUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(entityData)
}

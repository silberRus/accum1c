package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"inventory/internal/config"
	"net/http"
	"strings"

	"inventory/internal/service"
)

type Handler struct {
	service *service.Service
	config  *config.Config
}

type EndpointType int

const (
	UpdateEndpoint EndpointType = iota
	GetEndpoint
)

func NewHandler(service *service.Service, config *config.Config) *Handler {
	return &Handler{service: service, config: config}
}

func getEntityName(path string, w http.ResponseWriter, endpointType EndpointType) string {

	conf := config.GlobalConfig
	entityName := ""

	for _, entityConfig := range conf.DatabaseStructure {
		var currentEndpoint string
		switch endpointType {
		case UpdateEndpoint:
			currentEndpoint = entityConfig.UpdateEndpoint
		case GetEndpoint:
			currentEndpoint = entityConfig.GetEndpoint
		default:
			http.Error(w, "Invalid endpoint type", http.StatusBadRequest)
			return ""
		}
		if "/"+currentEndpoint == path {
			entityName = entityConfig.Entity
			break
		}
	}
	if entityName == "" {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
	return entityName
}

func (h *Handler) UpdateEntityHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("UpdateEntityHandler request: %v\n", r)

	entityName := getEntityName(r.URL.Path, w, UpdateEndpoint)
	if entityName == "" {
		return
	}
	fmt.Printf("Entity name: %s\n", entityName)

	conf := config.GlobalConfig
	entityConfig := conf.GetEntityConfig(entityName)
	fmt.Printf("Entity config: %v\n", entityConfig)

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

	fmt.Printf("GetEntityHandler request: %v\n", r)
	pathSegments := strings.SplitN(r.URL.Path, "/", 3)
	if len(pathSegments) < 2 {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		return
	}
	trimmedPath := "/" + pathSegments[1]

	entityName := getEntityName(trimmedPath, w, GetEndpoint)
	if entityName == "" {
		return
	}
	fmt.Printf("Entity name: %s\n", entityName)

	// Validate entity
	entityConfig := h.config.GetEntityConfig(entityName)
	if entityConfig == nil {
		http.Error(w, fmt.Sprintf("Invalid entity: %s not found in configuration", entityName), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	entityGUID := vars["guid"]
	fmt.Printf("Entity GUID: %s\n", entityGUID)

	entityData, err := h.service.GetEntity(entityName, entityGUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(entityData)
}

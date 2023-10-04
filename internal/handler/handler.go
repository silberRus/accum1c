package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"inventory/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	var request service.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.UpdateInventory(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	quantity, err := h.service.GetInventory(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"quantity": quantity}
	json.NewEncoder(w).Encode(response)
}

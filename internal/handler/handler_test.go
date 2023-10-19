package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"inventory/internal/config"
	"inventory/internal/repository"
	"inventory/internal/service"
)

func setup() (*Handler, *mux.Router) {

	cfg := config.GlobalConfig
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	repo := repository.NewRepository(rdb, cfg)
	serv := service.NewService(repo, cfg)
	hand := NewHandler(serv, cfg)

	r := mux.NewRouter()
	// Dynamically set routes based on the loaded configuration
	for _, entity := range cfg.DatabaseStructure {
		updateEndpoint := "/" + entity.UpdateEndpoint
		getEndpoint := "/" + entity.GetEndpoint + "/{guid}"

		r.HandleFunc(updateEndpoint, hand.UpdateEntityHandler).Methods("POST")
		r.HandleFunc(getEndpoint, hand.GetEntityHandler).Methods("GET")
	}
	return hand, r
}

func TestAddNewProduct(t *testing.T) {
	_, router := setup()

	product := map[string]interface{}{
		"guid": "test-guid",
		"code": "test-code",
		"name": "test-name",
	}
	payload, _ := json.Marshal(product)

	req, _ := http.NewRequest("POST", "/update_product", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d but got %d", http.StatusNoContent, rr.Code)
	}
}

func TestUpdateExistingProduct(t *testing.T) {
	_, router := setup()

	updatedProduct := map[string]interface{}{
		"guid": "test-guid",
		"name": "updated-name",
	}
	payload, _ := json.Marshal(updatedProduct)

	req, _ := http.NewRequest("POST", "/update_product", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d but got %d", http.StatusNoContent, rr.Code)
	}
}

func TestGetProductStore(t *testing.T) {
	_, router := setup()

	getProductResponse := sendRequest(router, "GET", "/get_product_store/workflow-guid", nil)
	if getProductResponse.Result().StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(getProductResponse.Body)
		t.Fatalf("Expected status code %d but got %d. Response body: %s", http.StatusOK, getProductResponse.Result().StatusCode, string(bodyBytes))
	}

	var product map[string]interface{}
	json.Unmarshal(getProductResponse.Body.Bytes(), &product)

	currentSum := product["sum"]
	println("currentSum", currentSum)
}

func TestUpdateStore(t *testing.T) {
	_, router := setup()

	getProductResponse := sendRequest(router, "GET", "/get_product/workflow-guid", nil)
	if getProductResponse.Result().StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(getProductResponse.Body)
		t.Fatalf("Expected status code %d but got %d. Response body: %s", http.StatusOK, getProductResponse.Result().StatusCode, string(bodyBytes))
	}

	var product map[string]interface{}
	json.Unmarshal(getProductResponse.Body.Bytes(), &product)
	//currentSum := product["sum"].(float64)
	currentSum := product["sum"]
	println(currentSum)
}

func TestProductWorkflow(t *testing.T) {
	_, router := setup()

	// 1. Добавить новый продукт с начальным количеством
	initialProduct := map[string]interface{}{
		"guid": "workflow-guid",
		"code": "workflow-code",
		"name": "workflow-name",
		"sum":  100.0,
	}
	initialPayload, _ := json.Marshal(initialProduct)
	addProductResponse := sendRequest(router, "POST", "/update_product", initialPayload)
	fmt.Printf("Add product response: %v\n", addProductResponse.Result())
	if addProductResponse.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status code %d but got %d", http.StatusNoContent, addProductResponse.Result().StatusCode)
	}

	// 2. Списать некоторое количество продукта
	deductPayload, _ := json.Marshal(map[string]interface{}{
		"guid": "workflow-guid",
		"sum":  -50.0,
	})
	deductResponse := sendRequest(router, "POST", "/update_product_store:workflow-guid", deductPayload) // используем новый ключ
	fmt.Printf("Deduct product response: %v\n", deductResponse.Result())
	if deductResponse.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status code %d but got %d", http.StatusNoContent, deductResponse.Result().StatusCode)
	}

	// 3. Получить текущее количество продукта
	getProductResponse := sendRequest(router, "GET", "/get_product:workflow-guid", nil) // используем новый ключ
	if getProductResponse.Result().StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(getProductResponse.Body)
		t.Fatalf("Expected status code %d but got %d. Response body: %s", http.StatusOK, getProductResponse.Result().StatusCode, string(bodyBytes))
	}

	var product map[string]interface{}
	json.Unmarshal(getProductResponse.Body.Bytes(), &product)
	currentSum, _ := product["sum"].(float64)

	// 4. Проверить текущее количество
	if currentSum != 50.0 {
		t.Fatalf("Expected quantity 50.0 but got %f", currentSum)
	}

	// 5. Попробовать списать больше продукта, чем доступно
	overDeductProduct := map[string]interface{}{
		"guid": "workflow-guid",
		"sum":  -100.0,
	}
	overDeductPayload, _ := json.Marshal(overDeductProduct)
	overDeductResponse := sendRequest(router, "POST", "/update_product:workflow-guid", overDeductPayload) // используем новый ключ
	if overDeductResponse.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code %d but got %d", http.StatusBadRequest, overDeductResponse.Result().StatusCode)
	}
}

func sendRequest(router *mux.Router, method, url string, payload []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

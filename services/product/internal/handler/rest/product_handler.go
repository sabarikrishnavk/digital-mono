package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/libs/metrics"
	"github.com/omni-compos/digital-mono/services/product/internal/service"
)

// ProductRESTHandler handles HTTP requests for products.
type ProductRESTHandler struct {
	service service.ProductService
	logger  logger.Logger
	metrics metrics.PrometheusMetrics
}

// NewProductRESTHandler creates a new ProductRESTHandler.
func NewProductRESTHandler(productService service.ProductService, log logger.Logger, promMetrics metrics.PrometheusMetrics) *ProductRESTHandler {
	return &ProductRESTHandler{service: productService, logger: log, metrics: promMetrics}
}

// RegisterRoutes registers product REST routes.
func (h *ProductRESTHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/products", h.CreateProductHandler).Methods(http.MethodPost)
	router.HandleFunc("/products/{id}", h.GetProductHandler).Methods(http.MethodGet)
}

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SKU         string `json:"sku"`
}

func (h *ProductRESTHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		h.metrics.IncRequestsTotal("create-product",  "rest",  strconv.Itoa(http.StatusBadRequest))
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req.Name, req.Description, req.SKU)
	if err != nil {
		h.logger.Error(err, "Failed to create product")
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		h.metrics.IncRequestsTotal(r.URL.Path,  "rest",strconv.Itoa(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
	h.metrics.IncRequestsTotal(r.URL.Path,  "rest",  strconv.Itoa(http.StatusCreated))
}

func (h *ProductRESTHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get product", http.StatusInternalServerError)
		h.metrics.IncRequestsTotal(r.URL.Path, "rest",  strconv.Itoa(http.StatusInternalServerError))
		return
	}
	if product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		h.metrics.IncRequestsTotal(r.URL.Path, "rest", strconv.Itoa(http.StatusNotFound))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
	h.metrics.IncRequestsTotal(r.URL.Path,  "rest",  strconv.Itoa(http.StatusOK))
}
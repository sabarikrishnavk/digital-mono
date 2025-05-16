package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	"github.com/omni-compos/digital-mono/libs/logger"
	commonMetrics "github.com/omni-compos/digital-mono/libs/metrics"
	model "github.com/omni-compos/digital-mono/services/seller/internal/domain"
	"github.com/omni-compos/digital-mono/services/seller/internal/service"
)

// SellerRESTHandler handles REST requests for sellers.
type SellerRESTHandler struct {
	service service.SellerService
	logger  logger.Logger
	metrics commonMetrics.PrometheusMetrics
}

// NewSellerRESTHandler creates a new SellerRESTHandler.
func NewSellerRESTHandler(service service.SellerService, logger logger.Logger, metrics commonMetrics.PrometheusMetrics) *SellerRESTHandler {
	return &SellerRESTHandler{
		service: service,
		logger:  logger,
		metrics: metrics,
	}
}

// RegisterRoutes registers the REST endpoints for sellers.
func (h *SellerRESTHandler) RegisterRoutes(router *mux.Router) {

	router.HandleFunc("/sellers", h.ListSellers).Methods(http.MethodGet) 
	router.HandleFunc("/sellers", h.CreateSeller).Methods(http.MethodPost) 
	router.HandleFunc("/sellers/{id}", h.GetSellerByID).Methods(http.MethodGet)
	router.HandleFunc("/sellers/{id}", h.UpdateSeller).Methods(http.MethodPut)
	router.HandleFunc("/sellers/{id}", h.DeleteSeller).Methods(http.MethodDelete) 

}

// CreateSeller handles POST /sellers
func (h *SellerRESTHandler) CreateSeller(w http.ResponseWriter, r *http.Request) {  
	// h.logger.Info("Entering CreateSeller handler", "method", r.Method, "path", r.URL.Path)
	h.metrics.IncRequestsTotal("create_seller", "rest")  
	timer := h.metrics.NewRequestDurationTimer("create_seller", "rest")
	defer timer.ObserveDuration() 

	var seller model.Seller
	if err := json.NewDecoder(r.Body).Decode(&seller); err != nil {
		h.logger.Error(err, "Failed to decode request body for CreateSeller")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
 
		h.metrics.IncResponsesTotal("create_seller", "rest", strconv.Itoa(http.StatusBadRequest))
		return
	} 
	// Get UserID from JWT claims in context 
	claims, ok := commonAuth.GetClaimsFromContext(r.Context())
	if !ok   {
		h.logger.Error(nil, "UserID not found in context for CreateSeller")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		h.metrics.IncResponsesTotal("create_seller", "rest", strconv.Itoa(http.StatusUnauthorized))
		return
	}

	createdSeller, err := h.service.CreateSeller(r.Context(), &seller, claims.UserID)
	if err != nil {
		h.logger.Error(err, "Failed to create seller via service")
		// More specific error handling could be added here (e.g., validation errors)
		http.Error(w, fmt.Sprintf("Failed to create seller: %v", err), http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("create_seller", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSeller)
	h.metrics.IncResponsesTotal("create_seller", "rest", strconv.Itoa(http.StatusCreated))
}

// GetSellerByID handles GET /sellers/{id}
func (h *SellerRESTHandler) GetSellerByID(w http.ResponseWriter, r *http.Request) {
	// h.logger.Info("Entering GetSellerByID handler", "method", r.Method, "path", r.URL.Path)
	h.metrics.IncRequestsTotal("get_seller_by_id", "rest")
	timer := h.metrics.NewRequestDurationTimer("get_seller_by_id", "rest")
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	id := vars["id"]

	seller, err := h.service.GetSellerByID(r.Context(), id)
	if err != nil {
		h.logger.Error(err, "Failed to get seller by ID via service", "seller_id", id)
		http.Error(w, fmt.Sprintf("Failed to retrieve seller: %v", err), http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("get_seller_by_id", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	if seller == nil {
		http.Error(w, "Seller not found", http.StatusNotFound)
		h.metrics.IncResponsesTotal("get_seller_by_id", "rest", strconv.Itoa(http.StatusNotFound))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seller)
	h.metrics.IncResponsesTotal("get_seller_by_id", "rest", strconv.Itoa(http.StatusOK))
}

// UpdateSeller handles PUT /sellers/{id}
func (h *SellerRESTHandler) UpdateSeller(w http.ResponseWriter, r *http.Request) {
	// h.logger.Info("Entering UpdateSeller handler", "method", r.Method, "path", r.URL.Path)
	h.metrics.IncRequestsTotal("update_seller", "rest")
	timer := h.metrics.NewRequestDurationTimer("update_seller", "rest")
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates model.Seller
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.logger.Error(err, "Failed to decode request body for UpdateSeller")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.metrics.IncResponsesTotal("update_seller", "rest", strconv.Itoa(http.StatusBadRequest))
		return
	}
 
	// Get UserID from JWT claims in context 
	claims, ok := commonAuth.GetClaimsFromContext(r.Context())
	if !ok   { 
		h.logger.Error(nil, "UserID not found in context for UpdateSeller")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		h.metrics.IncResponsesTotal("update_seller", "rest", strconv.Itoa(http.StatusUnauthorized))
		return
	}

	updatedSeller, err := h.service.UpdateSeller(r.Context(), id, &updates, claims.UserID)
	if err != nil {
		h.logger.Error(err, "Failed to update seller via service", "seller_id", id)
		// Check for specific errors like "not found"
		if err.Error() == fmt.Sprintf("seller with ID %s not found", id) { // Basic string match, improve with custom error types
			http.Error(w, "Seller not found", http.StatusNotFound)
			h.metrics.IncResponsesTotal("update_seller", "rest", strconv.Itoa(http.StatusNotFound))
			return
		}
		http.Error(w, fmt.Sprintf("Failed to update seller: %v", err), http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("update_seller", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSeller)
	h.metrics.IncResponsesTotal("update_seller", "rest", strconv.Itoa(http.StatusOK))
}

// DeleteSeller handles DELETE /sellers/{id}
func (h *SellerRESTHandler) DeleteSeller(w http.ResponseWriter, r *http.Request) {
	// h.logger.Info("Entering DeleteSeller handler", "method", r.Method, "path", r.URL.Path)
	h.metrics.IncRequestsTotal("delete_seller", "rest")
	timer := h.metrics.NewRequestDurationTimer("delete_seller", "rest")
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteSeller(r.Context(), id)
	if err != nil {
		h.logger.Error(err, "Failed to delete seller via service", "seller_id", id)
		// Check for specific errors like "not found"
		if err.Error() == fmt.Sprintf("seller with ID %s not found for delete", id) { // Basic string match, improve with custom error types
			http.Error(w, "Seller not found", http.StatusNotFound)
			h.metrics.IncResponsesTotal("delete_seller", "rest", strconv.Itoa(http.StatusNotFound))
			return
		}
		http.Error(w, fmt.Sprintf("Failed to delete seller: %v", err), http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("delete_seller", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content is typical for successful deletion
	h.metrics.IncResponsesTotal("delete_seller", "rest", strconv.Itoa(http.StatusNoContent))
}

// ListSellers handles GET /sellers
func (h *SellerRESTHandler) ListSellers(w http.ResponseWriter, r *http.Request) {
	// h.logger.Info("Entering ListSellers handler", "method", r.Method, "path", r.URL.Path)
	h.metrics.IncRequestsTotal("list_seller-s", "rest")
	timer := h.metrics.NewRequestDurationTimer("list_seller-s", "rest")
	defer timer.ObserveDuration()

	// Get pagination parameters from query string
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			h.metrics.IncResponsesTotal("list_sellers", "rest", strconv.Itoa(http.StatusBadRequest))
			return
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else {
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			h.metrics.IncResponsesTotal("list_sellers", "rest", strconv.Itoa(http.StatusBadRequest))
			return
		}
	}

	sellers, err := h.service.ListSellers(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error(err, "Failed to list sellers via service")
		http.Error(w, fmt.Sprintf("Failed to retrieve sellers: %v", err), http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("list_sellers", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sellers)
	h.metrics.IncResponsesTotal("list_sellers", "rest", strconv.Itoa(http.StatusOK))
}
package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/omni-compos/digital-mono/libs/auth"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/libs/metrics"
	"github.com/omni-compos/digital-mono/services/user/internal/service"
)

// UserRESTHandler handles HTTP requests for users.
type UserRESTHandler struct {
	service       service.UserService
	logger        logger.Logger
	metrics       metrics.PrometheusMetrics // Using the common metrics interface
	authenticator *auth.JWTAuthenticator
}

// NewUserRESTHandler creates a new UserRESTHandler.
func NewUserRESTHandler(userService service.UserService, log logger.Logger, promMetrics metrics.PrometheusMetrics, authenticator *auth.JWTAuthenticator) *UserRESTHandler {
	return &UserRESTHandler{service: userService, logger: log, metrics: promMetrics, authenticator: authenticator}
}

// RegisterRoutes registers user REST routes with the given router.
func (h *UserRESTHandler) RegisterRoutes(router *mux.Router) {
	// Example of using metrics wrapper, if your metrics lib provides one
	// createUserHandler := h.metrics.WrapHandler("create_user", http.HandlerFunc(h.CreateUserHandler))
	// getUserHandler := h.metrics.WrapHandler("get_user", http.HandlerFunc(h.GetUserHandler))

	router.HandleFunc("/users", h.CreateUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", h.GetUserHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", h.LoginHandler).Methods(http.MethodPost)
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	// Using Email as username for this example
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserRESTHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusBadRequest))
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		h.logger.Error(err, "Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusCreated))
}

func (h *UserRESTHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user, err := h.service.GetUser(r.Context(), id)
	if err != nil { // Handle not found specifically if service returns a specific error
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusInternalServerError))
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusNotFound))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusOK))
}

func (h *UserRESTHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusBadRequest))
		return
	}

	// --- Mock Authentication ---
	// In a real application, you would validate req.Email and req.Password against your user database.
	// For this example, we'll assume authentication is successful if email and password are provided.
	// We'll also fetch the user to get their ID and potentially roles.
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusUnauthorized)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusUnauthorized))
		return
	}

	// Mock: find user by email to get ID. If your service.GetUserByEmail exists.
	// For now, let's assume a mock user ID and roles.
	// user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	// if err != nil || user == nil {
	// 	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	//  h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusUnauthorized))
	// 	return
	// }
	// userID := user.ID
	userID := "mock_user_id_for_" + req.Email // Placeholder
	roles := []string{"user", "viewer"}       // Mock roles

	token, err := h.authenticator.GenerateToken(userID, roles, 24*time.Hour) // Token valid for 24 hours
	if err != nil {
		h.logger.Error(err, "Failed to generate token")
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
	h.metrics.IncRequestsTotal(r.URL.Path, r.Method, http.StatusText(http.StatusOK))
}
package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	commonAuth "github.com/omni-compos/digital-mono/libs/auth"
	"github.com/omni-compos/digital-mono/libs/logger"
	"github.com/omni-compos/digital-mono/libs/metrics"
	"github.com/omni-compos/digital-mono/services/user/internal/domain"
	"github.com/omni-compos/digital-mono/services/user/internal/service"
)

// UserRESTHandler handles HTTP requests for users.
type UserRESTHandler struct {
	service service.UserService
	jwtSecret    string
	logger  logger.Logger
	metrics metrics.PrometheusMetrics // Using the common metrics interface
}

// NewUserRESTHandler creates a new UserRESTHandler.
func NewUserRESTHandler(userService service.UserService, jwtSecret string, log logger.Logger, promMetrics metrics.PrometheusMetrics) *UserRESTHandler {
	return &UserRESTHandler{service: userService,jwtSecret: jwtSecret, logger: log, metrics: promMetrics}
}

// // RegisterRoutes registers user REST routes with the given router.
// func (h *UserRESTHandler) RegisterRoutes(router *mux.Router) {
// 	// Example of using metrics wrapper, if your metrics lib provides one
// 	// createUserHandler := h.metrics.WrapHandler("create_user", http.HandlerFunc(h.CreateUserHandler))
// 	// getUserHandler := h.metrics.WrapHandler("get_user", http.HandlerFunc(h.GetUserHandler))

// 	router.HandleFunc("/users", h.CreateUserHandler).Methods(http.MethodPost)
// 	router.HandleFunc("/users/{id}", h.GetUserHandler).Methods(http.MethodGet)
// }

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
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

// RegisterRoutes registers the REST endpoints for users.
func (h *UserRESTHandler) RegisterRoutes(r *mux.Router) {
	// Public route for login (does not require JWT middleware)
	// Public routes
}

	// API routes requiring JWT authentication middleware
	// The middleware should be applied to the router *before* these routes are registered
	// In main.go, we'll create a subrouter for authenticated routes.

// RegisterProtectedRoutes registers routes that require authentication.
// These should be registered on a router or subrouter that has the JWT middleware applied.
func (h *UserRESTHandler) RegisterProtectedRoutes(r *mux.Router) {

	r.HandleFunc("/users", h.CreateUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", h.GetUserHandler).Methods(http.MethodGet)
	// r.HandleFunc("/users", h.CreateUser).Methods("POST")
	// r.HandleFunc("/users/{id}", h.GetUserByID).Methods("GET")
	// r.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	// r.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
	// r.HandleFunc("/users", h.ListUsers).Methods("GET")
}

// Login handles POST /login requests.
func (h *UserRESTHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.metrics.IncRequestsTotal("login", "rest",strconv.Itoa(http.StatusOK))
	timer := h.metrics.NewRequestDurationTimer("login", "rest")
	defer timer.ObserveDuration()

	var req *domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(err, "Failed to decode login request body")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.metrics.IncResponsesTotal("login", "rest", strconv.Itoa(http.StatusBadRequest))
		return
	}

	// Authenticate user using the service layer
	user, err := h.service.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn(err,"Authentication failed", "email", req.Email, "error")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		h.metrics.IncResponsesTotal("login", "rest", strconv.Itoa(http.StatusUnauthorized))
		return
	}

	// Generate JWT token
	// Use a reasonable expiration time, e.g., 24 hours

	authenticator := commonAuth.NewJWTAuthenticator(h.jwtSecret)
	token, err := authenticator.GenerateToken(user.ID, user.Roles, 24*time.Hour)
	if err != nil {
		h.logger.Error(err, "Failed to generate JWT token for user", "userID", user.ID)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		h.metrics.IncResponsesTotal("login", "rest", strconv.Itoa(http.StatusInternalServerError))
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(domain.LoginResponse{Token: token})
	h.metrics.IncResponsesTotal("login", "rest", strconv.Itoa(http.StatusOK))
}
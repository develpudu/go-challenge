package handler

import (
	"encoding/json"
	"net/http"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/entity"
)

// Handles HTTP requests related to users
type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// Creates a new user handler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// Represents the request body for creating a user
type CreateUserRequest struct {
	Username string `json:"username"`
}

// Represents the response body for user-related operations
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Represents the request body for following a user
type FollowRequest struct {
	FollowedID string `json:"followed_id"`
}

// Registers the user routes
func (h *UserHandler) RegisterRoutes() {
	http.HandleFunc("/users", h.handleUsers)
	http.HandleFunc("/users/", h.handleUserByID)
	http.HandleFunc("/users/follow", h.handleFollow)
	http.HandleFunc("/users/unfollow", h.handleUnfollow)
}

// Handles requests to /users
func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createUser(w, r)
	case http.MethodGet:
		h.getUsers(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Handles requests to /users/{id}
func (h *UserHandler) handleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := r.URL.Path[len("/users/"):]
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.getUser(w, r, userID)
}

// Handles requests to /users/follow
func (h *UserHandler) handleFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.followUser(w, r)
}

// Handles requests to /users/unfollow
func (h *UserHandler) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.unfollowUser(w, r)
}

// Creates a new user
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "username is required"})
		return
	}

	// Create user
	user, err := h.userUseCase.CreateUser(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}

// Returns all users
func (h *UserHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	// Get all users
	users, err := h.userUseCase.GetAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:       user.ID,
			Username: user.Username,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Returns a specific user
func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Get user
	user, err := h.userUseCase.GetUser(userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}

// Makes a user follow another user
func (h *UserHandler) followUser(w http.ResponseWriter, r *http.Request) {
	// Get follower ID from header
	followerID := r.Header.Get("User-ID")
	if followerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User-ID header is required"})
		return
	}

	// Parse request body
	var req FollowRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FollowedID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "followed_id is required"})
		return
	}

	// Follow user
	err = h.userUseCase.FollowUser(followerID, req.FollowedID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err == entity.ErrCannotFollowSelf {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User followed successfully"})
}

// Makes a user unfollow another user
func (h *UserHandler) unfollowUser(w http.ResponseWriter, r *http.Request) {
	// Get follower ID from header
	followerID := r.Header.Get("User-ID")
	if followerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User-ID header is required"})
		return
	}

	// Parse request body
	var req FollowRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate request
	if req.FollowedID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "followed_id is required"})
		return
	}

	// Unfollow user
	err = h.userUseCase.UnfollowUser(followerID, req.FollowedID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User unfollowed successfully"})
}

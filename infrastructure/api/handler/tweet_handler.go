package handler

import (
	"encoding/json"
	"net/http"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/entity"
)

// Handles HTTP requests related to tweets
type TweetHandler struct {
	tweetUseCase *usecase.TweetUseCase
}

// Creates a new tweet handler
func NewTweetHandler(tweetUseCase *usecase.TweetUseCase) *TweetHandler {
	return &TweetHandler{
		tweetUseCase: tweetUseCase,
	}
}

// Represents the request body for creating a tweet
type CreateTweetRequest struct {
	Content string `json:"content"`
}

// Represents the response body for tweet-related operations
type TweetResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Registers the tweet routes
func (h *TweetHandler) RegisterRoutes() {
	http.HandleFunc("/tweets", h.handleTweets)
	http.HandleFunc("/tweets/", h.handleTweetByID)
	http.HandleFunc("/users/tweets", h.handleUserTweets)
	http.HandleFunc("/timeline", h.handleTimeline)
}

// Handles requests to /tweets
func (h *TweetHandler) handleTweets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTweet(w, r)
	case http.MethodGet:
		h.getAllTweets(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Handles requests to /tweets/{id}
func (h *TweetHandler) handleTweetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract tweet ID from URL path
	tweetID := r.URL.Path[len("/tweets/"):]
	if tweetID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.getTweet(w, r, tweetID)
}

// Handles requests to /users/tweets
func (h *TweetHandler) handleUserTweets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.getUserTweets(w, r)
}

// Handles requests to /timeline
func (h *TweetHandler) handleTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.getTimeline(w, r)
}

// Creates a new tweet
func (h *TweetHandler) createTweet(w http.ResponseWriter, r *http.Request) {
	// Get user ID from header
	userID := r.Header.Get("User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User-ID header is required"})
		return
	}

	// Parse request body
	var req CreateTweetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "content is required"})
		return
	}

	// Create tweet
	tweet, err := h.tweetUseCase.CreateTweet(userID, req.Content)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
			return
		} else if err == entity.ErrTweetTooLong {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "tweet exceeds character limit"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TweetResponse{
		ID:        tweet.ID,
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Returns all tweets
func (h *TweetHandler) getAllTweets(w http.ResponseWriter, r *http.Request) {
	// Get all tweets
	tweets, err := h.tweetUseCase.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]TweetResponse, len(tweets))
	for i, tweet := range tweets {
		response[i] = TweetResponse{
			ID:        tweet.ID,
			UserID:    tweet.UserID,
			Content:   tweet.Content,
			CreatedAt: tweet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Returns a specific tweet
func (h *TweetHandler) getTweet(w http.ResponseWriter, r *http.Request, tweetID string) {
	// Get tweet
	tweet, err := h.tweetUseCase.FindByID(tweetID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if tweet == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TweetResponse{
		ID:        tweet.ID,
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Returns all tweets by a specific user
func (h *TweetHandler) getUserTweets(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameter
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id query parameter is required"})
		return
	}

	// Get tweets by user
	tweets, err := h.tweetUseCase.GetTweetsByUser(userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]TweetResponse, len(tweets))
	for i, tweet := range tweets {
		response[i] = TweetResponse{
			ID:        tweet.ID,
			UserID:    tweet.UserID,
			Content:   tweet.Content,
			CreatedAt: tweet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Returns the timeline for a specific user
func (h *TweetHandler) getTimeline(w http.ResponseWriter, r *http.Request) {
	// Get user ID from header
	userID := r.Header.Get("User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User-ID header is required"})
		return
	}

	// Get timeline
	tweets, err := h.tweetUseCase.GetTimeline(userID)
	if err != nil {
		if err == entity.ErrUserNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]TweetResponse, len(tweets))
	for i, tweet := range tweets {
		response[i] = TweetResponse{
			ID:        tweet.ID,
			UserID:    tweet.UserID,
			Content:   tweet.Content,
			CreatedAt: tweet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

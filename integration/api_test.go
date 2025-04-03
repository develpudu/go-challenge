package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/infrastructure/api/handler"
	"github.com/develpudu/go-challenge/infrastructure/repository/memory"
)

// Returns a test API server
func setupTestAPI(t *testing.T) (http.Handler, *memory.UserRepository, *memory.TweetRepository) {
	// Reset DefaultServeMux for each test
	http.DefaultServeMux = new(http.ServeMux)

	// Initialize in-memory repositories
	userRepo := memory.NewUserRepository()
	tweetRepo := memory.NewTweetRepository(userRepo)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	tweetUseCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUseCase)
	tweetHandler := handler.NewTweetHandler(tweetUseCase)

	// Register routes
	userHandler.RegisterRoutes()
	tweetHandler.RegisterRoutes()

	return http.DefaultServeMux, userRepo, tweetRepo
}

func TestCreateAndGetUser(t *testing.T) {
	// Setup
	router, _, _ := setupTestAPI(t)

	// Create a user
	userPayload := map[string]string{"username": "testuser"}
	userJSON, _ := json.Marshal(userPayload)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Parse response to get user ID
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	userID, ok := response["id"].(string)
	if !ok {
		t.Fatal("Expected user ID in response")
	}

	// Get the created user
	req, _ = http.NewRequest("GET", "/users/"+userID, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify user data
	var userData map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &userData)

	if userData["username"] != "testuser" {
		t.Errorf("Expected username to be 'testuser', got %v", userData["username"])
	}
}

func TestCreateAndGetTweet(t *testing.T) {
	// Setup
	router, userRepo, _ := setupTestAPI(t)

	// Create a user first
	user := entity.NewUser("user123", "testuser")
	userRepo.Save(user)

	// Create a tweet
	tweetPayload := map[string]string{
		"content": "This is a test tweet",
	}
	tweetJSON, _ := json.Marshal(tweetPayload)

	req, _ := http.NewRequest("POST", "/tweets", bytes.NewBuffer(tweetJSON))
	req.Header.Set("Content-Type", "application/json")
	// Set User-ID header as required by the handler
	req.Header.Set("User-ID", user.ID)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Parse response to get tweet ID
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	tweetID, ok := response["id"].(string)
	if !ok {
		t.Fatal("Expected tweet ID in response")
	}

	// Get the created tweet
	req, _ = http.NewRequest("GET", "/tweets/"+tweetID, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify tweet data
	var tweetData map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &tweetData)

	if tweetData["content"] != "This is a test tweet" {
		t.Errorf("Expected content to be 'This is a test tweet', got %v", tweetData["content"])
	}

	if tweetData["user_id"] != user.ID {
		t.Errorf("Expected user_id to be %s, got %v", user.ID, tweetData["user_id"])
	}
}

func TestFollowUserAndGetTimeline(t *testing.T) {
	// Setup
	router, userRepo, tweetRepo := setupTestAPI(t)

	// Create two users
	follower := entity.NewUser("follower123", "follower")
	followed := entity.NewUser("followed123", "followed")
	userRepo.Save(follower)
	userRepo.Save(followed)

	// Create a tweet for the followed user
	tweet, _ := entity.NewTweet("tweet123", followed.ID, "Tweet from followed user")
	tweetRepo.Save(tweet)

	// Make follower follow followed
	followPayload := map[string]string{"followed_id": followed.ID}
	followJSON, _ := json.Marshal(followPayload)
	req, _ := http.NewRequest("POST", "/users/follow", bytes.NewBuffer(followJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-ID", follower.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Get follower's timeline
	req, _ = http.NewRequest("GET", "/timeline", nil)
	req.Header.Set("User-ID", follower.ID)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify timeline contains the tweet from followed user
	var timeline []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &timeline)

	if len(timeline) == 0 {
		t.Fatal("Expected timeline to contain tweets, but it was empty")
	}

	foundTweet := false
	for _, t := range timeline {
		if t["id"] == tweet.ID {
			foundTweet = true
			break
		}
	}

	if !foundTweet {
		t.Error("Expected timeline to contain the tweet from followed user, but it was not found")
	}
}

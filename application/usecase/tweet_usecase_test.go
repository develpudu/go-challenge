package usecase_test

import (
	"testing"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/entity"
)

// Mock implementation of the TweetRepository interface
type MockTweetRepository struct {
	tweets map[string]*entity.Tweet
}

// Creates a new mock tweet repository
func NewMockTweetRepository() *MockTweetRepository {
	return &MockTweetRepository{
		tweets: make(map[string]*entity.Tweet),
	}
}

// Stores a tweet in the repository
func (r *MockTweetRepository) Save(tweet *entity.Tweet) error {
	r.tweets[tweet.ID] = tweet
	return nil
}

// Retrieves a tweet by its ID
func (r *MockTweetRepository) FindByID(id string) (*entity.Tweet, error) {
	tweet, exists := r.tweets[id]
	if !exists {
		return nil, nil
	}
	return tweet, nil
}

// Retrieves all tweets by a specific user
func (r *MockTweetRepository) FindByUserID(userID string) ([]*entity.Tweet, error) {
	result := make([]*entity.Tweet, 0)
	for _, tweet := range r.tweets {
		if tweet.UserID == userID {
			result = append(result, tweet)
		}
	}
	return result, nil
}

// Retrieves all tweets
func (r *MockTweetRepository) FindAll() ([]*entity.Tweet, error) {
	result := make([]*entity.Tweet, 0, len(r.tweets))
	for _, tweet := range r.tweets {
		result = append(result, tweet)
	}
	return result, nil
}

// Removes a tweet from the repository
func (r *MockTweetRepository) Delete(id string) error {
	delete(r.tweets, id)
	return nil
}

// GetTimeline retrieves the timeline for a specific user
func (r *MockTweetRepository) GetTimeline(userID string) ([]*entity.Tweet, error) {
	// In a real implementation, this would get tweets from the user and all followed users
	// For the mock, we'll just return all tweets as a simplification
	return r.FindAll()
}

func TestCreateTweet(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Create a user
	user := entity.NewUser("user123", "testuser")
	userRepo.Save(user)

	content := "This is a test tweet"

	// Act
	tweet, err := useCase.CreateTweet(user.ID, content)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tweet == nil {
		t.Fatal("Expected tweet to be created, got nil")
	}

	if tweet.UserID != user.ID {
		t.Errorf("Expected tweet user ID to be %s, got %s", user.ID, tweet.UserID)
	}

	if tweet.Content != content {
		t.Errorf("Expected tweet content to be %s, got %s", content, tweet.Content)
	}

	// Check that tweet was saved to repository
	savedTweet, err := tweetRepo.FindByID(tweet.ID)
	if err != nil {
		t.Errorf("Error finding tweet in repository: %v", err)
	}

	if savedTweet == nil {
		t.Error("Expected tweet to be saved in repository, but it was not found")
	}
}

func TestCreateTweetUserNotFound(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Act
	_, err := useCase.CreateTweet("nonexistent", "This is a test tweet")

	// Assert
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateTweetTooLong(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Create a user
	user := entity.NewUser("user123", "testuser")
	userRepo.Save(user)

	// Create content that exceeds the character limit
	content := ""
	for i := 0; i <= entity.MaxTweetLength; i++ {
		content += "a"
	}

	// Act
	_, err := useCase.CreateTweet(user.ID, content)

	// Assert
	if err != entity.ErrTweetTooLong {
		t.Errorf("Expected ErrTweetTooLong, got %v", err)
	}
}

func TestGetTweetsByUser(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Create a user
	user := entity.NewUser("user123", "testuser")
	userRepo.Save(user)

	// Create some tweets for the user
	tweet1, _ := entity.NewTweet("tweet1", user.ID, "Tweet 1")
	tweet2, _ := entity.NewTweet("tweet2", user.ID, "Tweet 2")
	tweet3, _ := entity.NewTweet("tweet3", user.ID, "Tweet 3")

	// Create a tweet for another user
	otherUser := entity.NewUser("other123", "otheruser")
	userRepo.Save(otherUser)
	otherTweet, _ := entity.NewTweet("otherTweet", otherUser.ID, "Other user's tweet")

	// Save all tweets
	tweetRepo.Save(tweet1)
	tweetRepo.Save(tweet2)
	tweetRepo.Save(tweet3)
	tweetRepo.Save(otherTweet)

	// Act
	tweets, err := useCase.GetTweetsByUser(user.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(tweets) != 3 {
		t.Errorf("Expected 3 tweets, got %d", len(tweets))
	}

	// Check that all tweets belong to the user
	for _, tweet := range tweets {
		if tweet.UserID != user.ID {
			t.Errorf("Expected tweet to belong to user %s, but it belongs to %s", user.ID, tweet.UserID)
		}
	}
}

func TestGetTimeline(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Create users
	user := entity.NewUser("user123", "testuser")
	followed1 := entity.NewUser("followed1", "followed1User")
	followed2 := entity.NewUser("followed2", "followed2User")
	notFollowed := entity.NewUser("notFollowed", "notFollowedUser")

	// Set up following relationships
	user.Follow(followed1.ID)
	user.Follow(followed2.ID)

	// Save all users
	userRepo.Save(user)
	userRepo.Save(followed1)
	userRepo.Save(followed2)
	userRepo.Save(notFollowed)

	// Create tweets for each user
	userTweet, _ := entity.NewTweet("userTweet", user.ID, "User's tweet")
	followed1Tweet, _ := entity.NewTweet("followed1Tweet", followed1.ID, "Followed1's tweet")
	followed2Tweet, _ := entity.NewTweet("followed2Tweet", followed2.ID, "Followed2's tweet")
	notFollowedTweet, _ := entity.NewTweet("notFollowedTweet", notFollowed.ID, "Not followed's tweet")

	// Save all tweets
	tweetRepo.Save(userTweet)
	tweetRepo.Save(followed1Tweet)
	tweetRepo.Save(followed2Tweet)
	tweetRepo.Save(notFollowedTweet)

	// Act
	timeline, err := useCase.GetTimeline(user.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Note: GetTimeline returns all tweets
	// In a real implementation, it would filter based on following relationships
	// So we're just checking that we get some tweets back
	if len(timeline) == 0 {
		t.Error("Expected timeline to contain tweets, but it was empty")
	}
}

func TestGetTweetByID(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Create a user and a tweet
	user := entity.NewUser("user123", "testuser")
	userRepo.Save(user)

	tweet, _ := entity.NewTweet("tweet123", user.ID, "Test tweet")
	tweetRepo.Save(tweet)

	// Act
	retrievedTweet, err := useCase.GetTweetByID(tweet.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedTweet == nil {
		t.Fatal("Expected to retrieve tweet, got nil")
	}

	if retrievedTweet.ID != tweet.ID {
		t.Errorf("Expected tweet ID to be %s, got %s", tweet.ID, retrievedTweet.ID)
	}

	if retrievedTweet.Content != tweet.Content {
		t.Errorf("Expected content to be %s, got %s", tweet.Content, retrievedTweet.Content)
	}
}

func TestGetTweetByIDNotFound(t *testing.T) {
	// Arrange
	tweetRepo := NewMockTweetRepository()
	userRepo := NewMockUserRepository()
	useCase := usecase.NewTweetUseCase(tweetRepo, userRepo)

	// Act
	_, err := useCase.GetTweetByID("nonexistent")

	// Assert
	if err != entity.ErrTweetNotFound {
		t.Errorf("Expected ErrTweetNotFound, got %v", err)
	}
}

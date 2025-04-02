package entity_test

import (
	"strings"
	"testing"
	"time"

	"github.com/develpudu/go-challenge/domain/entity"
)

func TestNewTweet(t *testing.T) {
	// Arrange
	id := "tweet123"
	userID := "user456"
	content := "This is a test tweet"

	// Act
	tweet, err := entity.NewTweet(id, userID, content)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tweet.ID != id {
		t.Errorf("Expected tweet ID to be %s, got %s", id, tweet.ID)
	}

	if tweet.UserID != userID {
		t.Errorf("Expected user ID to be %s, got %s", userID, tweet.UserID)
	}

	if tweet.Content != content {
		t.Errorf("Expected content to be %s, got %s", content, tweet.Content)
	}

	// Check that CreatedAt is set to a time close to now
	now := time.Now()
	diff := now.Sub(tweet.CreatedAt)
	if diff > time.Second*5 || diff < -time.Second*5 {
		t.Errorf("Expected CreatedAt to be close to current time, but difference was %v", diff)
	}
}

func TestNewTweetTooLong(t *testing.T) {
	// Arrange
	id := "tweet123"
	userID := "user456"
	// Create a tweet that exceeds the character limit
	content := strings.Repeat("a", entity.MaxTweetLength+1)

	// Act
	tweet, err := entity.NewTweet(id, userID, content)

	// Assert
	if err != entity.ErrTweetTooLong {
		t.Errorf("Expected ErrTweetTooLong, got %v", err)
	}

	if tweet != nil {
		t.Errorf("Expected tweet to be nil when content is too long, got %v", tweet)
	}
}

func TestTweetIsValid(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Valid tweet",
			content:  "This is a valid tweet",
			expected: true,
		},
		{
			name:     "Empty tweet",
			content:  "",
			expected: true,
		},
		{
			name:     "Maximum length tweet",
			content:  strings.Repeat("a", entity.MaxTweetLength),
			expected: true,
		},
		{
			name:     "Too long tweet",
			content:  strings.Repeat("a", entity.MaxTweetLength+1),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a tweet directly to bypass the NewTweet validation
			tweet := &entity.Tweet{
				ID:        "tweet123",
				UserID:    "user456",
				Content:   tc.content,
				CreatedAt: time.Now(),
			}

			// Act
			result := tweet.IsValid()

			// Assert
			if result != tc.expected {
				t.Errorf("Expected IsValid to return %v, got %v", tc.expected, result)
			}
		})
	}
}

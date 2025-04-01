package repository

import (
	"github.com/develpudu/go-challenge/domain/entity"
)

// Defines the interface for tweet data operations
type TweetRepository interface {
	// Stores a tweet in the repository
	Save(tweet *entity.Tweet) error

	// Retrieves a tweet by its ID
	FindByID(id string) (*entity.Tweet, error)

	// Retrieves all tweets by a specific user
	FindByUserID(userID string) ([]*entity.Tweet, error)

	// Retrieves all tweets
	FindAll() ([]*entity.Tweet, error)

	// Removes a tweet from the repository
	Delete(id string) error

	// Retrieves tweets from users that a specific user follows
	// ordered by creation time (newest first)
	GetTimeline(userID string) ([]*entity.Tweet, error)
}

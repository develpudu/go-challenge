package entity

import (
	"time"
)

// Defines the maximum number of characters allowed in a tweet
const MaxTweetLength = 280

// Tweet in the microblogging platform
type Tweet struct {
	ID        string
	UserID    string
	Content   string
	CreatedAt time.Time
}

// Creates a new tweet with the given parameters
// Returns an error if the content exceeds the character limit
func NewTweet(id, userID, content string) (*Tweet, error) {
	// Validate tweet length
	if len(content) > MaxTweetLength {
		return nil, ErrTweetTooLong
	}

	return &Tweet{
		ID:        id,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

// Checks if the tweet is valid (within character limit)
func (t *Tweet) IsValid() bool {
	return len(t.Content) <= MaxTweetLength
}

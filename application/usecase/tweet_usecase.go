package usecase

import (
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/domain/repository"
	"github.com/google/uuid"
)

// Implements the tweet use cases
type TweetUseCase struct {
	tweetRepository repository.TweetRepository
	userRepository  repository.UserRepository
}

// Creates a new tweet use case
func NewTweetUseCase(
	tweetRepository repository.TweetRepository,
	userRepository repository.UserRepository,
) *TweetUseCase {
	return &TweetUseCase{
		tweetRepository: tweetRepository,
		userRepository:  userRepository,
	}
}

// Creates a new tweet for a user
func (uc *TweetUseCase) CreateTweet(userID, content string) (*entity.Tweet, error) {
	// Check if user exists
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	// Generate a unique ID for the tweet
	tweetID := uuid.New().String()

	// Create a new tweet
	tweet, err := entity.NewTweet(tweetID, userID, content)
	if err != nil {
		return nil, err
	}

	// Save the tweet
	err = uc.tweetRepository.Save(tweet)
	if err != nil {
		return nil, err
	}

	return tweet, nil
}

// Retrieves all tweets by a specific user
func (uc *TweetUseCase) GetTweetsByUser(userID string) ([]*entity.Tweet, error) {
	// Check if user exists
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	// Get tweets by user ID
	return uc.tweetRepository.FindByUserID(userID)
}

// Retrieves the timeline for a specific user
// The timeline includes tweets from users that the user follows and their own tweets
func (uc *TweetUseCase) GetTimeline(userID string) ([]*entity.Tweet, error) {
	// Check if user exists
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	// Get timeline
	return uc.tweetRepository.GetTimeline(userID)
}

// Retrieves all tweets from the repository
func (uc *TweetUseCase) GetAllTweets() ([]*entity.Tweet, error) {
	return uc.tweetRepository.FindAll()
}

// Retrieves a specific tweet by its ID
func (uc *TweetUseCase) GetTweetByID(tweetID string) (*entity.Tweet, error) {
	tweet, err := uc.tweetRepository.FindByID(tweetID)
	if err != nil {
		return nil, err
	}
	if tweet == nil {
		return nil, entity.ErrTweetNotFound
	}
	return tweet, nil
}

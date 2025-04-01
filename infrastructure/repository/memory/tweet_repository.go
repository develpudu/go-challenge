package memory

import (
	"sort"
	"sync"

	"github.com/develpudu/go-challenge/domain/entity"
)

// Implements the tweet repository interface with an in-memory storage
type TweetRepository struct {
	tweets       map[string]*entity.Tweet   // Map of tweet ID to tweet
	userTweets   map[string][]*entity.Tweet // Map of user ID to their tweets
	userTimeline map[string][]*entity.Tweet // Cache of user timelines for optimization
	userRepo     *UserRepository
	mutex        sync.RWMutex
}

// Creates a new in-memory tweet repository
func NewTweetRepository(userRepo *UserRepository) *TweetRepository {
	return &TweetRepository{
		tweets:       make(map[string]*entity.Tweet),
		userTweets:   make(map[string][]*entity.Tweet),
		userTimeline: make(map[string][]*entity.Tweet),
		userRepo:     userRepo,
	}
}

// Stores a tweet in the repository
func (r *TweetRepository) Save(tweet *entity.Tweet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Store the tweet
	r.tweets[tweet.ID] = tweet

	// Add to user tweets
	r.userTweets[tweet.UserID] = append(r.userTweets[tweet.UserID], tweet)

	// Invalidate timelines that include this user's tweets
	// This is a simple approach; in a real system, we would use a more sophisticated cache invalidation strategy
	r.invalidateTimelines(tweet.UserID)

	return nil
}

// Retrieves a tweet by its ID
func (r *TweetRepository) FindByID(id string) (*entity.Tweet, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tweet, exists := r.tweets[id]
	if !exists {
		return nil, nil // Return nil, nil when tweet not found as per interface contract
	}

	return tweet, nil
}

// Retrieves all tweets by a specific user
func (r *TweetRepository) FindByUserID(userID string) ([]*entity.Tweet, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tweets, exists := r.userTweets[userID]
	if !exists {
		return []*entity.Tweet{}, nil // Return empty slice when no tweets found
	}

	// Sort tweets by creation time (newest first)
	sortedTweets := make([]*entity.Tweet, len(tweets))
	copy(sortedTweets, tweets)
	sort.Slice(sortedTweets, func(i, j int) bool {
		return sortedTweets[i].CreatedAt.After(sortedTweets[j].CreatedAt)
	})

	return sortedTweets, nil
}

// Retrieves all tweets
func (r *TweetRepository) FindAll() ([]*entity.Tweet, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tweets := make([]*entity.Tweet, 0, len(r.tweets))
	for _, tweet := range r.tweets {
		tweets = append(tweets, tweet)
	}

	// Sort tweets by creation time (newest first)
	sort.Slice(tweets, func(i, j int) bool {
		return tweets[i].CreatedAt.After(tweets[j].CreatedAt)
	})

	return tweets, nil
}

// Removes a tweet from the repository
func (r *TweetRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if tweet exists
	tweet, exists := r.tweets[id]
	if !exists {
		return entity.ErrTweetNotFound
	}

	// Get user ID for timeline invalidation
	userID := tweet.UserID

	// Remove from tweets map
	delete(r.tweets, id)

	// Remove from user tweets
	tweets := r.userTweets[userID]
	for i, t := range tweets {
		if t.ID == id {
			// Remove tweet from slice
			r.userTweets[userID] = append(tweets[:i], tweets[i+1:]...)
			break
		}
	}

	// Invalidate timelines
	r.invalidateTimelines(userID)

	return nil
}

// Retrieves tweets from users that a specific user follows
// ordered by creation time (newest first)
func (r *TweetRepository) GetTimeline(userID string) ([]*entity.Tweet, error) {
	r.mutex.RLock()

	// Check if we have a cached timeline
	cachedTimeline, exists := r.userTimeline[userID]
	if exists {
		r.mutex.RUnlock()
		return cachedTimeline, nil
	}

	// No cached timeline, we need to build it
	r.mutex.RUnlock()

	// Get the user
	user, err := r.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	// Get the IDs of users that this user follows
	followingIDs := user.GetFollowing()

	// Add the user's own ID to include their tweets in the timeline
	followingIDs = append(followingIDs, userID)

	// Lock for writing as we'll update the cache
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Collect tweets from followed users and the user themselves
	timeline := make([]*entity.Tweet, 0)
	for _, followedID := range followingIDs {
		if tweets, exists := r.userTweets[followedID]; exists {
			timeline = append(timeline, tweets...)
		}
	}

	// Sort timeline by creation time (newest first)
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].CreatedAt.After(timeline[j].CreatedAt)
	})

	// Cache the timeline
	r.userTimeline[userID] = timeline

	return timeline, nil
}

// Invalidates all timelines that include tweets from the specified user
func (r *TweetRepository) invalidateTimelines(userID string) {
	// In a real system, we would use a more sophisticated approach
	// For simplicity, we'll just clear all timelines
	r.userTimeline = make(map[string][]*entity.Tweet)
}

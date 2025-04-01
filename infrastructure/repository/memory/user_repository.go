package memory

import (
	"sync"

	"github.com/develpudu/go-challenge/domain/entity"
)

// Implements the user repository interface with an in-memory storage
type UserRepository struct {
	users map[string]*entity.User
	mutex sync.RWMutex
}

// Creates a new in-memory user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*entity.User),
	}
}

// Stores a user in the repository
func (r *UserRepository) Save(user *entity.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Store a copy of the user to prevent external modifications
	r.users[user.ID] = user
	return nil
}

// Retrieves a user by their ID
func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, nil // Return nil, nil when user not found as per interface contract
	}

	return user, nil
}

// Retrieves all users
func (r *UserRepository) FindAll() ([]*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

// Updates an existing user
func (r *UserRepository) Update(user *entity.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	_, exists := r.users[user.ID]
	if !exists {
		return entity.ErrUserNotFound
	}

	// Update user
	r.users[user.ID] = user
	return nil
}

// Removes a user from the repository
func (r *UserRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	_, exists := r.users[id]
	if !exists {
		return entity.ErrUserNotFound
	}

	// Delete user
	delete(r.users, id)
	return nil
}

// Retrieves all users that follow a specific user
func (r *UserRepository) FindFollowers(userID string) ([]*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	followers := make([]*entity.User, 0)

	// Iterate through all users and check if they follow the specified user
	for _, user := range r.users {
		if user.IsFollowing(userID) {
			followers = append(followers, user)
		}
	}

	return followers, nil
}

// Retrieves all users that a specific user follows
func (r *UserRepository) FindFollowing(userID string) ([]*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Find the user
	user, exists := r.users[userID]
	if !exists {
		return nil, entity.ErrUserNotFound
	}

	// Get the IDs of users that this user follows
	followingIDs := user.GetFollowing()

	// Create a slice to store the following users
	following := make([]*entity.User, 0, len(followingIDs))

	// Get the user objects for each following ID
	for _, id := range followingIDs {
		if followedUser, exists := r.users[id]; exists {
			following = append(following, followedUser)
		}
	}

	return following, nil
}

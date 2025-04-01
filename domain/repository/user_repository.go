package repository

import (
	"github.com/develpudu/go-challenge/domain/entity"
)

// Defines the interface for user data operations
type UserRepository interface {
	// Stores a user in the repository
	Save(user *entity.User) error

	// Retrieves a user by their ID
	FindByID(id string) (*entity.User, error)

	// Retrieves all users
	FindAll() ([]*entity.User, error)

	// Udates an existing user
	Update(user *entity.User) error

	// Removes a user from the repository
	Delete(id string) error

	// Retrieves all users that follow a specific user
	FindFollowers(userID string) ([]*entity.User, error)

	// Retrieves all users that a specific user follows
	FindFollowing(userID string) ([]*entity.User, error)
}

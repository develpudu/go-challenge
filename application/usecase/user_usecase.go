package usecase

import (
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/domain/repository"
	"github.com/google/uuid"
)

// Implements the user use cases
type UserUseCase struct {
	userRepository repository.UserRepository
}

// Creates a new user use case
func NewUserUseCase(userRepository repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepository: userRepository,
	}
}

// Creates a new user
func (uc *UserUseCase) CreateUser(username string) (*entity.User, error) {
	// Generate a unique ID for the user
	userID := uuid.New().String()

	// Create a new user
	user := entity.NewUser(userID, username)

	// Save the user
	err := uc.userRepository.Save(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Retrieves a user by ID
func (uc *UserUseCase) GetUser(userID string) (*entity.User, error) {
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	return user, nil
}

// Makes a user follow another user
func (uc *UserUseCase) FollowUser(followerID, followedID string) error {
	// Check if both users exist
	follower, err := uc.userRepository.FindByID(followerID)
	if err != nil {
		return err
	}
	if follower == nil {
		return entity.ErrUserNotFound
	}

	followed, err := uc.userRepository.FindByID(followedID)
	if err != nil {
		return err
	}
	if followed == nil {
		return entity.ErrUserNotFound
	}

	// Make follower follow followed
	err = follower.Follow(followedID)
	if err != nil {
		return err
	}

	// Update follower in repository
	return uc.userRepository.Update(follower)
}

// Makes a user unfollow another user
func (uc *UserUseCase) UnfollowUser(followerID, followedID string) error {
	// Check if follower exists
	follower, err := uc.userRepository.FindByID(followerID)
	if err != nil {
		return err
	}
	if follower == nil {
		return entity.ErrUserNotFound
	}

	// Make follower unfollow followed
	follower.Unfollow(followedID)

	// Update follower in repository
	return uc.userRepository.Update(follower)
}

// Retrieves all users that follow a specific user
func (uc *UserUseCase) GetFollowers(userID string) ([]*entity.User, error) {
	// Check if user exists
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	return uc.userRepository.FindFollowers(userID)
}

// Retrieves all users that a specific user follows
func (uc *UserUseCase) GetFollowing(userID string) ([]*entity.User, error) {
	// Check if user exists
	user, err := uc.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	return uc.userRepository.FindFollowing(userID)
}

// Retrieves all users from the repository
func (uc *UserUseCase) GetAllUsers() ([]*entity.User, error) {
	return uc.userRepository.FindAll()
}

package usecase_test

import (
	"context"
	"testing"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/infrastructure/cache"
)

// Mock implementation of the UserRepository interface
type MockUserRepository struct {
	users map[string]*entity.User
}

// Creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
	}
}

// Stores a user in the repository
func (r *MockUserRepository) Save(user *entity.User) error {
	r.users[user.ID] = user
	return nil
}

// Retrieves a user by their ID
func (r *MockUserRepository) FindByID(id string) (*entity.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// Retrieves all users
func (r *MockUserRepository) FindAll() ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

// Updates an existing user
func (r *MockUserRepository) Update(user *entity.User) error {
	r.users[user.ID] = user
	return nil
}

// Removes a user from the repository
func (r *MockUserRepository) Delete(id string) error {
	delete(r.users, id)
	return nil
}

// Retrieves all users that follow a specific user
func (r *MockUserRepository) FindFollowers(userID string) ([]*entity.User, error) {
	followers := make([]*entity.User, 0)
	for _, user := range r.users {
		if user.IsFollowing(userID) {
			followers = append(followers, user)
		}
	}
	return followers, nil
}

// Retrieves all users that a specific user follows
func (r *MockUserRepository) FindFollowing(userID string) ([]*entity.User, error) {
	user, exists := r.users[userID]
	if !exists {
		return nil, nil
	}

	following := make([]*entity.User, 0, len(user.Following))
	for followedID := range user.Following {
		if followedUser, exists := r.users[followedID]; exists {
			following = append(following, followedUser)
		}
	}

	return following, nil
}

// Mock implementation of TimelineCache interface
type MockTimelineCache struct{}

func (m *MockTimelineCache) GetTimeline(ctx context.Context, userID string) ([]*entity.Tweet, bool, error) {
	return nil, false, nil // Always cache miss
}
func (m *MockTimelineCache) SetTimeline(ctx context.Context, userID string, timeline []*entity.Tweet) error {
	return nil // Do nothing
}
func (m *MockTimelineCache) InvalidateTimeline(ctx context.Context, userID string) error {
	return nil // Do nothing
}

// Compile-time check
var _ cache.TimelineCache = (*MockTimelineCache)(nil)

func TestCreateUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache
	username := "testuser"

	// Act
	user, err := useCase.CreateUser(username)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be created, got nil")
	}

	if user.Username != username {
		t.Errorf("Expected username to be %s, got %s", username, user.Username)
	}

	// Check that user was saved to repository
	savedUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Errorf("Error finding user in repository: %v", err)
	}

	if savedUser == nil {
		t.Error("Expected user to be saved in repository, but it was not found")
	}
}

func TestGetUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create a user to retrieve
	user := entity.NewUser("user123", "testuser")
	repo.Save(user)

	// Act
	retrievedUser, err := useCase.GetUser(user.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedUser == nil {
		t.Fatal("Expected to retrieve user, got nil")
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID to be %s, got %s", user.ID, retrievedUser.ID)
	}

	if retrievedUser.Username != user.Username {
		t.Errorf("Expected username to be %s, got %s", user.Username, retrievedUser.Username)
	}
}

func TestGetUserNotFound(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Act
	_, err := useCase.GetUser("nonexistent")

	// Assert
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestFollowUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create two users
	follower := entity.NewUser("follower", "followerUser")
	followed := entity.NewUser("followed", "followedUser")

	repo.Save(follower)
	repo.Save(followed)

	// Act
	err := useCase.FollowUser(follower.ID, followed.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that follower is now following followed
	updatedFollower, _ := repo.FindByID(follower.ID)
	if !updatedFollower.IsFollowing(followed.ID) {
		t.Error("Expected follower to be following followed, but IsFollowing returned false")
	}
}

func TestFollowUserSelf(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create a user
	user := entity.NewUser("user123", "testuser")
	repo.Save(user)

	// Act
	err := useCase.FollowUser(user.ID, user.ID)

	// Assert
	if err != entity.ErrCannotFollowSelf {
		t.Errorf("Expected ErrCannotFollowSelf, got %v", err)
	}
}

func TestUnfollowUser(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create two users
	follower := entity.NewUser("follower", "followerUser")
	followed := entity.NewUser("followed", "followedUser")

	// Make follower follow followed
	follower.Follow(followed.ID)

	repo.Save(follower)
	repo.Save(followed)

	// Act
	err := useCase.UnfollowUser(follower.ID, followed.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that follower is no longer following followed
	updatedFollower, _ := repo.FindByID(follower.ID)
	if updatedFollower.IsFollowing(followed.ID) {
		t.Error("Expected follower to not be following followed after unfollowing, but IsFollowing returned true")
	}
}

func TestGetFollowers(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create users
	user := entity.NewUser("user", "mainUser")
	follower1 := entity.NewUser("follower1", "follower1User")
	follower2 := entity.NewUser("follower2", "follower2User")
	nonFollower := entity.NewUser("nonFollower", "nonFollowerUser")

	// Set up following relationships
	follower1.Follow(user.ID)
	follower2.Follow(user.ID)

	// Save all users
	repo.Save(user)
	repo.Save(follower1)
	repo.Save(follower2)
	repo.Save(nonFollower)

	// Act
	followers, err := useCase.GetFollowers(user.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(followers) != 2 {
		t.Errorf("Expected 2 followers, got %d", len(followers))
	}

	// Check that both followers are in the result
	followerIDs := make(map[string]bool)
	for _, f := range followers {
		followerIDs[f.ID] = true
	}

	if !followerIDs[follower1.ID] {
		t.Errorf("Expected follower1 to be in the followers list, but it was not found")
	}

	if !followerIDs[follower2.ID] {
		t.Errorf("Expected follower2 to be in the followers list, but it was not found")
	}

	if followerIDs[nonFollower.ID] {
		t.Errorf("Expected nonFollower to not be in the followers list, but it was found")
	}
}

func TestGetFollowing(t *testing.T) {
	// Arrange
	repo := NewMockUserRepository()
	cache := &MockTimelineCache{}                  // Use mock cache
	useCase := usecase.NewUserUseCase(repo, cache) // Pass cache

	// Create users
	user := entity.NewUser("user", "mainUser")
	followed1 := entity.NewUser("followed1", "followed1User")
	followed2 := entity.NewUser("followed2", "followed2User")
	notFollowed := entity.NewUser("notFollowed", "notFollowedUser")

	// Set up following relationships
	user.Follow(followed1.ID)
	user.Follow(followed2.ID)

	// Save all users
	repo.Save(user)
	repo.Save(followed1)
	repo.Save(followed2)
	repo.Save(notFollowed)

	// Act
	following, err := useCase.GetFollowing(user.ID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(following) != 2 {
		t.Errorf("Expected 2 following, got %d", len(following))
	}

	// Check that both followed users are in the result
	followedIDs := make(map[string]bool)
	for _, f := range following {
		followedIDs[f.ID] = true
	}

	if !followedIDs[followed1.ID] {
		t.Errorf("Expected followed1 to be in the following list, but it was not found")
	}

	if !followedIDs[followed2.ID] {
		t.Errorf("Expected followed2 to be in the following list, but it was not found")
	}

	if followedIDs[notFollowed.ID] {
		t.Errorf("Expected notFollowed to not be in the following list, but it was found")
	}
}

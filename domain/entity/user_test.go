package entity_test

import (
	"testing"

	"github.com/develpudu/go-challenge/domain/entity"
)

func TestNewUser(t *testing.T) {
	// Arrange
	id := "user123"
	username := "testuser"

	// Act
	user := entity.NewUser(id, username)

	// Assert
	if user.ID != id {
		t.Errorf("Expected user ID to be %s, got %s", id, user.ID)
	}

	if user.Username != username {
		t.Errorf("Expected username to be %s, got %s", username, user.Username)
	}

	if len(user.Following) != 0 {
		t.Errorf("Expected new user to have 0 followings, got %d", len(user.Following))
	}
}

func TestUserFollow(t *testing.T) {
	// Arrange
	user := entity.NewUser("user1", "testuser1")
	otherUserID := "user2"

	// Act
	err := user.Follow(otherUserID)

	// Assert
	if err != nil {
		t.Errorf("Expected no error when following another user, got %v", err)
	}

	if !user.IsFollowing(otherUserID) {
		t.Errorf("Expected user to be following %s, but IsFollowing returned false", otherUserID)
	}
}

func TestUserFollowSelf(t *testing.T) {
	// Arrange
	userID := "user1"
	user := entity.NewUser(userID, "testuser1")

	// Act
	err := user.Follow(userID)

	// Assert
	if err != entity.ErrCannotFollowSelf {
		t.Errorf("Expected ErrCannotFollowSelf when following self, got %v", err)
	}

	if user.IsFollowing(userID) {
		t.Errorf("Expected user not to be following self, but IsFollowing returned true")
	}
}

func TestUserUnfollow(t *testing.T) {
	// Arrange
	user := entity.NewUser("user1", "testuser1")
	otherUserID := "user2"
	_ = user.Follow(otherUserID) // First follow the user

	// Act
	user.Unfollow(otherUserID)

	// Assert
	if user.IsFollowing(otherUserID) {
		t.Errorf("Expected user not to be following %s after unfollowing, but IsFollowing returned true", otherUserID)
	}
}

func TestGetFollowing(t *testing.T) {
	// Arrange
	user := entity.NewUser("user1", "testuser1")
	followedIDs := []string{"user2", "user3", "user4"}

	// Follow multiple users
	for _, id := range followedIDs {
		_ = user.Follow(id)
	}

	// Act
	following := user.GetFollowing()

	// Assert
	if len(following) != len(followedIDs) {
		t.Errorf("Expected %d followings, got %d", len(followedIDs), len(following))
	}

	// Check that all followed IDs are in the result
	for _, id := range followedIDs {
		found := false
		for _, followingID := range following {
			if followingID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find user ID %s in following list, but it was not found", id)
		}
	}
}

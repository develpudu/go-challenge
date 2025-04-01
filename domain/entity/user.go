package entity

// User in the microblogging platform
type User struct {
	ID        string
	Username  string
	Following map[string]bool // Map of user IDs that this user follows
}

// Creates a new user with the given ID and username
func NewUser(id, username string) *User {
	return &User{
		ID:        id,
		Username:  username,
		Following: make(map[string]bool),
	}
}

// Makes the user follow another user
func (u *User) Follow(userID string) error {
	// User cannot follow themselves
	if u.ID == userID {
		return ErrCannotFollowSelf
	}

	u.Following[userID] = true
	return nil
}

// Makes the user unfollow another user
func (u *User) Unfollow(userID string) {
	delete(u.Following, userID)
}

// Checks if the user is following another user
func (u *User) IsFollowing(userID string) bool {
	_, following := u.Following[userID]
	return following
}

// Returns a slice of user IDs that this user follows
func (u *User) GetFollowing() []string {
	following := make([]string, 0, len(u.Following))
	for id := range u.Following {
		following = append(following, id)
	}
	return following
}

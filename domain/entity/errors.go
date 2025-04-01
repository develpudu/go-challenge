package entity

import "errors"

// Domain errors
var (
	// Returned when a user tries to follow themselves
	ErrCannotFollowSelf = errors.New("user cannot follow themselves")

	// Returned when a tweet exceeds the character limit
	ErrTweetTooLong = errors.New("tweet exceeds character limit")

	// Returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// Returned when a tweet is not found
	ErrTweetNotFound = errors.New("tweet not found")
)

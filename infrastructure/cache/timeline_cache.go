package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/go-redis/redis/v8"
)

const (
	// Default TTL for cached timelines
	defaultTimelineTTL = 5 * time.Minute
	// Key prefix for timeline cache entries in Redis
	timelineKeyPrefix = "timeline:"
)

// TimelineCache defines the interface for caching user timelines.
type TimelineCache interface {
	// GetTimeline retrieves a cached timeline for a user.
	// Returns the timeline, a boolean indicating if found, and an error.
	GetTimeline(ctx context.Context, userID string) ([]*entity.Tweet, bool, error)

	// SetTimeline caches a timeline for a user with a default TTL.
	SetTimeline(ctx context.Context, userID string, timeline []*entity.Tweet) error

	// InvalidateTimeline removes a cached timeline for a user.
	InvalidateTimeline(ctx context.Context, userID string) error
}

// RedisTimelineCache implements TimelineCache using Redis.
type RedisTimelineCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisTimelineCache creates a new Redis timeline cache client.
// It reads the Redis endpoint from the REDIS_ENDPOINT environment variable.
func NewRedisTimelineCache(ctx context.Context) (*RedisTimelineCache, error) {
	redisEndpoint := os.Getenv("REDIS_ENDPOINT")
	if redisEndpoint == "" {
		return nil, errors.New("REDIS_ENDPOINT environment variable not set")
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisEndpoint,
		// Add other options like Password, DB if needed
	})

	// Ping the server to ensure connectivity
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close() // Close client if ping fails
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", redisEndpoint, err)
	}

	// Use slog for info message
	slog.InfoContext(ctx, "Connected to Redis", "endpoint", redisEndpoint)
	return &RedisTimelineCache{
		client: client,
		ttl:    defaultTimelineTTL,
	}, nil
}

// generateKey creates the Redis key for a user's timeline.
func (c *RedisTimelineCache) generateKey(userID string) string {
	return timelineKeyPrefix + userID
}

// GetTimeline retrieves a cached timeline for a user from Redis.
func (c *RedisTimelineCache) GetTimeline(ctx context.Context, userID string) ([]*entity.Tweet, bool, error) {
	key := c.generateKey(userID)
	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		slog.DebugContext(ctx, "Timeline cache miss", "userID", userID)
		return nil, false, nil // Cache miss
	}
	if err != nil {
		// Log the error but return it so the caller can potentially fetch from DB
		slog.ErrorContext(ctx, "Failed to get timeline from Redis", "userID", userID, "error", err)
		return nil, false, fmt.Errorf("failed to get timeline for user %s from Redis: %w", userID, err)
	}

	// Deserialize the timeline from JSON
	var timeline []*entity.Tweet
	if err := json.Unmarshal([]byte(val), &timeline); err != nil {
		// Use slog for warning
		slog.WarnContext(ctx, "Failed to unmarshal cached timeline, invalidating entry", "userID", userID, "error", err)
		_ = c.InvalidateTimeline(ctx, userID)
		return nil, false, fmt.Errorf("failed to unmarshal cached timeline for user %s: %w", userID, err)
	}

	slog.DebugContext(ctx, "Timeline cache hit", "userID", userID)
	return timeline, true, nil // Cache hit
}

// SetTimeline caches a timeline for a user in Redis.
func (c *RedisTimelineCache) SetTimeline(ctx context.Context, userID string, timeline []*entity.Tweet) error {
	key := c.generateKey(userID)

	// Serialize the timeline to JSON
	val, err := json.Marshal(timeline)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal timeline for caching", "userID", userID, "error", err)
		return fmt.Errorf("failed to marshal timeline for caching for user %s: %w", userID, err)
	}

	err = c.client.Set(ctx, key, val, c.ttl).Err()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to set timeline cache in Redis", "userID", userID, "error", err)
		return fmt.Errorf("failed to set timeline cache for user %s in Redis: %w", userID, err)
	}
	slog.DebugContext(ctx, "Successfully set timeline cache", "userID", userID, "ttl", c.ttl)
	return nil
}

// InvalidateTimeline removes a cached timeline for a user from Redis.
func (c *RedisTimelineCache) InvalidateTimeline(ctx context.Context, userID string) error {
	key := c.generateKey(userID)
	err := c.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil { // Ignore error if key didn't exist
		slog.ErrorContext(ctx, "Failed to invalidate timeline cache in Redis", "userID", userID, "error", err)
		return fmt.Errorf("failed to invalidate timeline cache for user %s in Redis: %w", userID, err)
	}
	slog.DebugContext(ctx, "Invalidated timeline cache", "userID", userID)
	return nil
}

// Close closes the Redis client connection.
func (c *RedisTimelineCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Compile-time check to ensure RedisTimelineCache implements TimelineCache
var _ TimelineCache = (*RedisTimelineCache)(nil)

package dynamodb

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/domain/repository"
	"github.com/develpudu/go-challenge/infrastructure/cache"
	"golang.org/x/sync/errgroup"
)

const (
	// Assumed name for the GSI on UserID. Must match the IaC template.
	userIDIndexName = "UserIDIndex"
)

// DynamoDBTweetRepository implements the TweetRepository interface using AWS DynamoDB.
type DynamoDBTweetRepository struct {
	client    *dynamodb.Client
	tableName string
	userRepo  repository.UserRepository // Needed for GetTimeline
	cache     cache.TimelineCache       // Added cache field
}

// dynamoDBTweet is a helper struct for marshalling/unmarshalling Tweet data.
type dynamoDBTweet struct {
	ID        string `dynamodbav:"ID"`
	UserID    string `dynamodbav:"UserID"`
	Content   string `dynamodbav:"Content"`
	CreatedAt string `dynamodbav:"CreatedAt"` // Store as ISO 8601 string
}

// NewDynamoDBTweetRepository creates a new DynamoDB tweet repository.
// It now accepts a TimelineCache instance.
func NewDynamoDBTweetRepository(cfg aws.Config, tableName string, userRepo repository.UserRepository, timelineCache cache.TimelineCache) *DynamoDBTweetRepository {
	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBTweetRepository{
		client:    client,
		tableName: tableName,
		userRepo:  userRepo,
		cache:     timelineCache, // Store the cache instance
	}
}

// toDynamoDBTweet converts an entity.Tweet to its DynamoDB representation.
func toDynamoDBTweet(tweet *entity.Tweet) (*dynamoDBTweet, error) {
	return &dynamoDBTweet{
		ID:        tweet.ID,
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt.Format(time.RFC3339Nano),
	}, nil
}

// fromDynamoDBTweet converts a DynamoDB item representation to an entity.Tweet.
func fromDynamoDBTweet(ddbTweet *dynamoDBTweet) (*entity.Tweet, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, ddbTweet.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CreatedAt timestamp '%s': %w", ddbTweet.CreatedAt, err)
	}
	return &entity.Tweet{
		ID:        ddbTweet.ID,
		UserID:    ddbTweet.UserID,
		Content:   ddbTweet.Content,
		CreatedAt: createdAt,
	}, nil
}

// Save stores a tweet in the DynamoDB table.
// It also invalidates the author's timeline cache.
func (r *DynamoDBTweetRepository) Save(tweet *entity.Tweet) error {
	ctx := context.Background() // Use a background context for now
	ddbTweet, err := toDynamoDBTweet(tweet)
	if err != nil {
		return fmt.Errorf("failed to convert tweet to DynamoDB format: %w", err)
	}

	av, err := attributevalue.MarshalMap(ddbTweet)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet to attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to save tweet to DynamoDB: %w", err)
	}

	// Invalidate timeline cache for the author
	if r.cache != nil {
		if err := r.cache.InvalidateTimeline(ctx, tweet.UserID); err != nil {
			// Log the invalidation error but don't fail the Save operation
			fmt.Printf("WARN: Failed to invalidate timeline cache for user %s after saving tweet %s: %v\n", tweet.UserID, tweet.ID, err)
		}
	} else {
		fmt.Println("WARN: Timeline cache is nil, skipping invalidation.")
	}

	// TODO: Implement more robust invalidation for followers' timelines
	// This might involve fetching followers and invalidating each one, potentially asynchronously.

	return nil
}

// FindByID retrieves a tweet by its ID from DynamoDB.
func (r *DynamoDBTweetRepository) FindByID(id string) (*entity.Tweet, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"ID": id})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key for FindByID: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	result, err := r.client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get tweet from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Tweet not found
	}

	var ddbTweet dynamoDBTweet
	err = attributevalue.UnmarshalMap(result.Item, &ddbTweet)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tweet from DynamoDB: %w", err)
	}

	return fromDynamoDBTweet(&ddbTweet)
}

// queryTweetsByUserIDWithContext performs a query against the UserIDIndex GSI, propagating context.
func (r *DynamoDBTweetRepository) queryTweetsByUserIDWithContext(ctx context.Context, userID string) ([]*entity.Tweet, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String(userIDIndexName),
		KeyConditionExpression: aws.String("UserID = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
		// ScanIndexForward: aws.Bool(false), // To sort by sort key descending if one exists
	}

	paginator := dynamodb.NewQueryPaginator(r.client, input)

	tweets := make([]*entity.Tweet, 0)
	for paginator.HasMorePages() {
		// Pass the context to NextPage
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query tweets page for user %s: %w", userID, err)
		}

		var pageTweets []dynamoDBTweet
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageTweets)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tweets page for user %s: %w", userID, err)
		}

		for _, ddbTweet := range pageTweets {
			entityTweet, err := fromDynamoDBTweet(&ddbTweet)
			if err != nil {
				// Log error but potentially continue?
				fmt.Printf("WARN: Failed to convert tweet %s for user %s: %v\n", ddbTweet.ID, userID, err)
				continue
			}
			tweets = append(tweets, entityTweet)
		}
	}

	return tweets, nil
}

// FindByUserID retrieves all tweets by a specific user using a GSI.
func (r *DynamoDBTweetRepository) FindByUserID(userID string) ([]*entity.Tweet, error) {
	// Ensure user exists? The use case layer already does this.
	// Use the new function with a background context for non-timeline calls
	return r.queryTweetsByUserIDWithContext(context.Background(), userID)
}

// FindAll retrieves all tweets from DynamoDB.
// WARNING: This uses Scan, which is inefficient for large tables. Consider alternatives in production.
func (r *DynamoDBTweetRepository) FindAll() ([]*entity.Tweet, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	paginator := dynamodb.NewScanPaginator(r.client, input)

	tweets := make([]*entity.Tweet, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to scan tweets page: %w", err)
		}

		var pageTweets []dynamoDBTweet
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageTweets)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal scanned tweets page: %w", err)
		}

		for _, ddbTweet := range pageTweets {
			entityTweet, err := fromDynamoDBTweet(&ddbTweet)
			if err != nil {
				fmt.Printf("WARN: Failed to convert scanned tweet %s: %v\n", ddbTweet.ID, err)
				continue
			}
			tweets = append(tweets, entityTweet)
		}
	}

	return tweets, nil
}

// Delete removes a tweet from the DynamoDB table.
// It also invalidates the author's timeline cache.
func (r *DynamoDBTweetRepository) Delete(id string) error {
	ctx := context.Background() // Use a background context for now

	// First, we need to get the tweet to find the author ID for invalidation
	tweet, err := r.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to find tweet %s before deleting: %w", id, err)
	}
	if tweet == nil {
		return entity.ErrTweetNotFound
	}
	authorID := tweet.UserID

	// Proceed with deletion
	key, err := attributevalue.MarshalMap(map[string]string{"ID": id})
	if err != nil {
		return fmt.Errorf("failed to marshal key for delete: %w", err)
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	_, err = r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete tweet %s from DynamoDB: %w", id, err)
	}

	// Invalidate timeline cache for the author
	if r.cache != nil {
		if err := r.cache.InvalidateTimeline(ctx, authorID); err != nil {
			fmt.Printf("WARN: Failed to invalidate timeline cache for user %s after deleting tweet %s: %v\n", authorID, id, err)
		}
	} else {
		fmt.Println("WARN: Timeline cache is nil, skipping invalidation.")
	}

	// TODO: Implement more robust invalidation for followers' timelines

	return nil
}

// GetTimeline retrieves tweets from the user and users they follow.
// It first checks the cache, then queries DynamoDB, stores in cache on miss.
func (r *DynamoDBTweetRepository) GetTimeline(userID string) ([]*entity.Tweet, error) {
	ctx := context.Background() // Use a background context for now

	// 1. Check cache first
	if r.cache != nil {
		cachedTimeline, found, err := r.cache.GetTimeline(ctx, userID)
		if err != nil {
			// Log error but proceed to fetch from DB
			fmt.Printf("WARN: Failed to get timeline for user %s from cache: %v\n", userID, err)
		}
		if found {
			fmt.Printf("Cache hit for timeline: %s\n", userID)
			return cachedTimeline, nil
		}
		fmt.Printf("Cache miss for timeline: %s\n", userID)
	} else {
		fmt.Println("WARN: Timeline cache is nil, cannot check cache.")
	}

	// 2. Cache miss or cache unavailable, fetch from DB (existing logic)
	if r.userRepo == nil {
		return nil, fmt.Errorf("userRepository is nil, cannot GetTimeline")
	}
	user, err := r.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s for timeline: %w", userID, err)
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	idsToFetch := make([]string, 0, len(user.Following)+1)
	idsToFetch = append(idsToFetch, userID)
	for followedID := range user.Following {
		idsToFetch = append(idsToFetch, followedID)
	}

	var allTweets []*entity.Tweet
	var mu sync.Mutex
	// Use errgroup with the same background context for now
	g, queryCtx := errgroup.WithContext(ctx)

	for _, id := range idsToFetch {
		fetchID := id
		g.Go(func() error {
			userTweets, err := r.queryTweetsByUserIDWithContext(queryCtx, fetchID)
			if err != nil {
				return fmt.Errorf("failed to get tweets for user %s during timeline fetch: %w", fetchID, err)
			}
			mu.Lock()
			allTweets = append(allTweets, userTweets...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(allTweets, func(i, j int) bool {
		return allTweets[i].CreatedAt.After(allTweets[j].CreatedAt)
	})

	// 3. Store fetched result in cache
	if r.cache != nil {
		if err := r.cache.SetTimeline(ctx, userID, allTweets); err != nil {
			// Log error but return the fetched data anyway
			fmt.Printf("WARN: Failed to set timeline cache for user %s: %v\n", userID, err)
		}
	}

	return allTweets, nil
}

// Compile-time check to ensure DynamoDBTweetRepository implements TweetRepository
var _ repository.TweetRepository = (*DynamoDBTweetRepository)(nil)

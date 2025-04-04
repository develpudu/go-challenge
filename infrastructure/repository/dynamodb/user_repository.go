package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/develpudu/go-challenge/domain/entity"
	"github.com/develpudu/go-challenge/domain/repository"
)

// DynamoDBUserRepository implements the UserRepository interface using AWS DynamoDB.
type DynamoDBUserRepository struct {
	client    *dynamodb.Client
	tableName string
}

// dynamoDBUser is a helper struct for marshalling/unmarshalling User data to/from DynamoDB.
// We store Following as a String Set (SS).
type dynamoDBUser struct {
	ID        string   `dynamodbav:"ID"`
	Username  string   `dynamodbav:"Username"`
	Following []string `dynamodbav:"Following,stringset,omitempty"` // Store keys of the map as a string set
}

// NewDynamoDBUserRepository creates a new DynamoDB user repository.
func NewDynamoDBUserRepository(cfg aws.Config, tableName string) *DynamoDBUserRepository {
	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBUserRepository{
		client:    client,
		tableName: tableName,
	}
}

// toDynamoDBUser converts an entity.User to its DynamoDB representation.
func toDynamoDBUser(user *entity.User) (*dynamoDBUser, error) {
	followingSet := make([]string, 0, len(user.Following))
	for id := range user.Following {
		followingSet = append(followingSet, id)
	}
	return &dynamoDBUser{
		ID:        user.ID,
		Username:  user.Username,
		Following: followingSet,
	}, nil
}

// fromDynamoDBUser converts a DynamoDB item representation to an entity.User.
func fromDynamoDBUser(ddbUser *dynamoDBUser) *entity.User {
	followingMap := make(map[string]bool)
	for _, id := range ddbUser.Following {
		followingMap[id] = true
	}
	return &entity.User{
		ID:        ddbUser.ID,
		Username:  ddbUser.Username,
		Following: followingMap,
	}
}

// Save stores a user in the DynamoDB table.
func (r *DynamoDBUserRepository) Save(user *entity.User) error {
	ddbUser, err := toDynamoDBUser(user)
	if err != nil {
		return fmt.Errorf("failed to convert user to DynamoDB format: %w", err)
	}

	av, err := attributevalue.MarshalMap(ddbUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user to attribute values: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}

	_, err = r.client.PutItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to save user to DynamoDB: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their ID from DynamoDB.
func (r *DynamoDBUserRepository) FindByID(id string) (*entity.User, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"ID": id})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	result, err := r.client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, nil // User not found, return nil, nil as per interface contract
	}

	var ddbUser dynamoDBUser
	err = attributevalue.UnmarshalMap(result.Item, &ddbUser)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from DynamoDB: %w", err)
	}

	return fromDynamoDBUser(&ddbUser), nil
}

// FindAll retrieves all users from DynamoDB.
// WARNING: This uses Scan, which is inefficient for large tables. Consider alternatives in production.
func (r *DynamoDBUserRepository) FindAll() ([]*entity.User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	paginator := dynamodb.NewScanPaginator(r.client, input)

	users := make([]*entity.User, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to scan users page from DynamoDB: %w", err)
		}

		var pageUsers []dynamoDBUser
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageUsers)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal users page from DynamoDB: %w", err)
		}

		for _, ddbUser := range pageUsers {
			users = append(users, fromDynamoDBUser(&ddbUser))
		}
	}

	return users, nil
}

// Update updates an existing user in DynamoDB.
// This implementation replaces the entire item. More granular updates are possible.
func (r *DynamoDBUserRepository) Update(user *entity.User) error {
	// For simplicity, we use PutItem which acts as an upsert.
	// A stricter Update would first check if the item exists using a ConditionExpression.
	return r.Save(user)
}

// Delete removes a user from the DynamoDB table.
func (r *DynamoDBUserRepository) Delete(id string) error {
	key, err := attributevalue.MarshalMap(map[string]string{"ID": id})
	if err != nil {
		return fmt.Errorf("failed to marshal key for delete: %w", err)
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
		// Optional: Add ConditionExpression to ensure item exists before deleting
		// ConditionExpression: aws.String("attribute_exists(ID)"),
	}

	_, err = r.client.DeleteItem(context.TODO(), input)
	if err != nil {
		// Consider handling specific errors, e.g., ConditionalCheckFailedException
		return fmt.Errorf("failed to delete user from DynamoDB: %w", err)
	}
	return nil
}

// FindFollowers retrieves all users that follow a specific user.
// WARNING: This uses Scan with a filter, which is very inefficient for large tables.
// A GSI on the 'Following' attribute might be needed for production use cases.
func (r *DynamoDBUserRepository) FindFollowers(userID string) ([]*entity.User, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("contains(Following, :userID)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	}

	paginator := dynamodb.NewScanPaginator(r.client, input)

	followers := make([]*entity.User, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to scan followers page from DynamoDB: %w", err)
		}

		var pageUsers []dynamoDBUser
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageUsers)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal followers page from DynamoDB: %w", err)
		}

		for _, ddbUser := range pageUsers {
			// We need to filter out the user themselves if they accidentally follow themselves in the data
			// (though the domain logic prevents this)
			if ddbUser.ID != userID {
				followers = append(followers, fromDynamoDBUser(&ddbUser))
			}
		}
	}

	return followers, nil
}

// FindFollowing retrieves all users that a specific user follows.
func (r *DynamoDBUserRepository) FindFollowing(userID string) ([]*entity.User, error) {
	user, err := r.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s for finding following: %w", userID, err)
	}
	if user == nil {
		return nil, entity.ErrUserNotFound // Or return empty list? Interface contract unclear. Assuming error.
	}

	if len(user.Following) == 0 {
		return []*entity.User{}, nil
	}

	// Prepare keys for BatchGetItem
	keys := make([]map[string]types.AttributeValue, 0, len(user.Following))
	for followedID := range user.Following {
		key, err := attributevalue.MarshalMap(map[string]string{"ID": followedID})
		if err != nil {
			// Log this error, but potentially continue? Or fail fast?
			return nil, fmt.Errorf("failed to marshal key for followed user %s: %w", followedID, err)
		}
		keys = append(keys, key)
	}

	// BatchGetItem has a limit of 100 items per request. Handle pagination if needed.
	// For simplicity, assuming less than 100 followings here.
	if len(keys) > 100 {
		// TODO: Implement pagination for BatchGetItem if > 100 keys
		return nil, fmt.Errorf("finding more than 100 followed users is not implemented yet")
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			r.tableName: {
				Keys: keys,
			},
		},
	}

	result, err := r.client.BatchGetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get following users from DynamoDB: %w", err)
	}

	followingUsers := make([]*entity.User, 0, len(result.Responses[r.tableName]))
	var ddbUsers []dynamoDBUser
	err = attributevalue.UnmarshalListOfMaps(result.Responses[r.tableName], &ddbUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal following users from DynamoDB: %w", err)
	}

	for _, ddbUser := range ddbUsers {
		followingUsers = append(followingUsers, fromDynamoDBUser(&ddbUser))
	}

	// TODO: Handle UnprocessedKeys if any

	return followingUsers, nil
}

// Compile-time check to ensure DynamoDBUserRepository implements UserRepository
var _ repository.UserRepository = (*DynamoDBUserRepository)(nil)

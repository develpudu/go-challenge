package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/domain/repository"
	"github.com/develpudu/go-challenge/infrastructure/api/handler"
	cacheRepo "github.com/develpudu/go-challenge/infrastructure/cache"
	dynamodbRepo "github.com/develpudu/go-challenge/infrastructure/repository/dynamodb"
	memoryRepo "github.com/develpudu/go-challenge/infrastructure/repository/memory"
)

// Use the correct type name
var httpAdapter *httpadapter.HandlerAdapter

// Main function - Entry point of the application
func main() {
	fmt.Println("Starting Microblogging Platform...")

	var userRepository repository.UserRepository
	var tweetRepository repository.TweetRepository
	var timelineCache cacheRepo.TimelineCache

	// Check command-line arguments to decide which repository implementation to use
	runMode := "local"
	if len(os.Args) > 1 && os.Args[1] == "aws" {
		runMode = "lambda"
	}

	if runMode == "lambda" {
		fmt.Println("Initializing DynamoDB repositories and Redis cache...")
		ctx := context.Background()

		// Initialize Redis Cache
		redisCache, err := cacheRepo.NewRedisTimelineCache(ctx)
		if err != nil {
			fmt.Printf("WARN: Failed to initialize Redis timeline cache: %v. Proceeding without cache.\n", err)
			timelineCache = nil
		} else {
			timelineCache = redisCache
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("unable to load AWS SDK config, %v", err)
		}

		// Use hardcoded table names
		usersTableName := "users"
		tweetsTableName := "tweets"
		fmt.Printf("Using DynamoDB tables: Users='%s', Tweets='%s'\n", usersTableName, tweetsTableName)

		// Initialize DynamoDB repositories
		ddbUserRepo := dynamodbRepo.NewDynamoDBUserRepository(cfg, usersTableName)
		userRepository = ddbUserRepo
		tweetRepository = dynamodbRepo.NewDynamoDBTweetRepository(cfg, tweetsTableName, ddbUserRepo, timelineCache)

	} else {
		fmt.Println("Initializing in-memory repositories...")
		timelineCache = nil
		// Initialize in-memory repositories
		memUserRepo := memoryRepo.NewUserRepository()
		userRepository = memUserRepo
		tweetRepository = memoryRepo.NewTweetRepository(memUserRepo)
	}

	// Initialize use cases (inject cache into UserUseCase)
	userUseCase := usecase.NewUserUseCase(userRepository, timelineCache)
	tweetUseCase := usecase.NewTweetUseCase(tweetRepository, userRepository)

	// Initialize API handlers
	userHandler := handler.NewUserHandler(userUseCase)
	tweetHandler := handler.NewTweetHandler(tweetUseCase)

	// Register routes
	userHandler.RegisterRoutes()
	tweetHandler.RegisterRoutes()

	// Run based on the determined mode
	if runMode == "lambda" {
		fmt.Println("Running in Lambda mode")
		// Use httpadapter to wrap the existing http.Handler (DefaultServeMux)
		httpAdapter = httpadapter.New(http.DefaultServeMux)
		lambda.Start(LambdaHandler)
	} else {
		fmt.Println("Running in local/Docker mode")
		// Start HTTP server
		log.Println("Server starting on port 8080")
		log.Println("API is ready to use!")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

// LambdaHandler proxies requests to the httpAdapter
func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return httpAdapter.ProxyWithContext(ctx, req)
}

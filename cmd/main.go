package main

import (
	"context"
	"log/slog"
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
	// Setup structured logging
	logLevel := slog.LevelInfo // Default level
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger) // Set as default logger

	slog.Info("Starting Microblogging Platform...")

	var userRepository repository.UserRepository
	var tweetRepository repository.TweetRepository
	var timelineCache cacheRepo.TimelineCache

	// Check command-line arguments to decide which repository implementation to use
	runMode := "local"
	if len(os.Args) > 1 && os.Args[1] == "aws" {
		runMode = "lambda"
	}
	slog.Info("Determined run mode", "mode", runMode)

	if runMode == "lambda" {
		slog.Info("Initializing DynamoDB repositories and Redis cache...")
		ctx := context.Background()

		// Initialize Redis Cache
		redisCache, err := cacheRepo.NewRedisTimelineCache(ctx)
		if err != nil {
			// Use structured logging for warnings
			slog.Warn("Failed to initialize Redis timeline cache. Proceeding without cache.", "error", err)
			timelineCache = nil
		} else {
			slog.Info("Redis timeline cache initialized.")
			timelineCache = redisCache
		}

		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			// Use Fatalf equivalent with slog: log error then exit
			slog.Error("unable to load AWS SDK config", "error", err)
			os.Exit(1)
		}

		// Use hardcoded table names
		usersTableName := "users"
		tweetsTableName := "tweets"
		slog.Info("Using DynamoDB tables", "usersTable", usersTableName, "tweetsTable", tweetsTableName)

		// Initialize DynamoDB repositories
		ddbUserRepo := dynamodbRepo.NewDynamoDBUserRepository(cfg, usersTableName)
		userRepository = ddbUserRepo
		tweetRepository = dynamodbRepo.NewDynamoDBTweetRepository(cfg, tweetsTableName, ddbUserRepo, timelineCache)

	} else {
		slog.Info("Initializing in-memory repositories...")
		timelineCache = nil
		// Initialize in-memory repositories
		memUserRepo := memoryRepo.NewUserRepository()
		userRepository = memUserRepo
		tweetRepository = memoryRepo.NewTweetRepository(memUserRepo)
	}

	slog.Info("Initializing use cases...")
	// Initialize use cases (inject cache into UserUseCase)
	userUseCase := usecase.NewUserUseCase(userRepository, timelineCache)
	tweetUseCase := usecase.NewTweetUseCase(tweetRepository, userRepository)

	// Initialize API handlers
	userHandler := handler.NewUserHandler(userUseCase)
	tweetHandler := handler.NewTweetHandler(tweetUseCase)

	slog.Info("Initializing API handlers and registering routes...")
	// Register routes
	userHandler.RegisterRoutes()
	tweetHandler.RegisterRoutes()

	// Run based on the determined mode
	if runMode == "lambda" {
		slog.Info("Starting Lambda handler")
		// Use httpadapter to wrap the existing http.Handler (DefaultServeMux)
		httpAdapter = httpadapter.New(http.DefaultServeMux)
		lambda.Start(LambdaHandler)
	} else {
		slog.Info("Starting HTTP server", "port", 8080)
		// Start HTTP server
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}
}

// LambdaHandler proxies requests to the httpAdapter
func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Add basic request logging
	slog.InfoContext(ctx, "Received Lambda request", "method", req.HTTPMethod, "path", req.Path, "requestID", req.RequestContext.RequestID)
	// Note: Consider adding more details like User-Agent, source IP if needed

	response, err := httpAdapter.ProxyWithContext(ctx, req)

	// Log response status
	if err != nil {
		slog.ErrorContext(ctx, "Lambda handler error", "error", err)
	} else {
		slog.InfoContext(ctx, "Sending Lambda response", "statusCode", response.StatusCode)
	}
	return response, err
}

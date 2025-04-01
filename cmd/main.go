package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/develpudu/go-challenge/application/usecase"
	"github.com/develpudu/go-challenge/infrastructure/api/handler"
	"github.com/develpudu/go-challenge/infrastructure/repository/memory"
)

// Main function - Entry point of the application
func main() {
	fmt.Println("Starting Microblogging Platform...")

	// Initialize repositories
	userRepository := memory.NewUserRepository()
	tweetRepository := memory.NewTweetRepository(userRepository)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepository)
	tweetUseCase := usecase.NewTweetUseCase(tweetRepository, userRepository)

	// Initialize API handlers
	userHandler := handler.NewUserHandler(userUseCase)
	tweetHandler := handler.NewTweetHandler(tweetUseCase)

	// Register routes
	userHandler.RegisterRoutes()
	tweetHandler.RegisterRoutes()

	// Start HTTP server
	log.Println("Server starting on port 8080")
	log.Println("API is ready to use!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

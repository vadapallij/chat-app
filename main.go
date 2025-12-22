package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"chat-app/handlers"
	"chat-app/services"
	"chat-app/workflows"

	"github.com/dbos-inc/dbos-transact-golang/dbos"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to PostgreSQL for app data
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL database")

	// Initialize Anthropic service
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}
	anthropicService := services.NewAnthropicService(apiKey)

	// Initialize workflows
	chatWorkflows := workflows.NewChatWorkflows(db, anthropicService)

	// Initialize DBOS context for durable workflows
	dbosCtx, err := dbos.NewDBOSContext(context.Background(), dbos.Config{
		DatabaseURL: dbURL,
		AppName:     "chat-app",
	})
	if err != nil {
		log.Fatalf("Failed to initialize DBOS: %v", err)
	}

	// Register workflows with DBOS (MUST be before Launch)
	dbos.RegisterWorkflow(dbosCtx, chatWorkflows.SendMessageWorkflow)
	dbos.RegisterWorkflow(dbosCtx, chatWorkflows.CreateConversationWorkflow)
	dbos.RegisterWorkflow(dbosCtx, chatWorkflows.DeleteConversationWorkflow)

	// Launch DBOS (starts workflow recovery)
	if err := dbos.Launch(dbosCtx); err != nil {
		log.Fatalf("Failed to launch DBOS: %v", err)
	}
	defer dbos.Shutdown(dbosCtx, 5*time.Second)
	log.Println("DBOS initialized - durable workflows enabled")

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(db, anthropicService, dbosCtx, chatWorkflows)

	// Setup Gin router
	router := gin.Default()

	// Enable CORS for local development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Conversation routes
		api.POST("/conversations", chatHandler.CreateConversation)
		api.GET("/conversations", chatHandler.ListConversations)
		api.GET("/conversations/:id", chatHandler.GetConversation)
		api.DELETE("/conversations/:id", chatHandler.DeleteConversation)

		// Message routes
		api.POST("/conversations/:id/messages", chatHandler.SendMessage)
		api.GET("/conversations/:id/messages", chatHandler.GetMessages)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "dbos": "enabled"})
	})

	// Serve static files
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

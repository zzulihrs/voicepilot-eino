package main

import (
	"log"

	"github.com/deca/voicepilot-eino/internal/config"
	"github.com/deca/voicepilot-eino/internal/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
	log.Printf("Server will listen on port: %s", config.AppConfig.Port)
	log.Printf("Safe mode: %v", config.AppConfig.EnableSafeMode)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Create handler
	h := handler.NewHandler()

	// Routes
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Voice interaction
		api.POST("/voice", h.VoiceInteraction)

		// Text interaction
		api.POST("/text", h.TextInteraction)

		// Audio upload (for testing)
		api.POST("/upload", h.UploadAudio)
	}

	// Static files
	r.GET("/static/audio/:filename", h.ServeAudio)

	// Start server
	addr := ":" + config.AppConfig.Port
	log.Printf("Starting VoicePilot-Eino server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

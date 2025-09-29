package main

import (
	"fmt"
	"log"
	"os"
  "github.com/gin-contrib/cors"
	"github.com/Mastermind730/igc-admin-backend/handlers"
	"github.com/Mastermind730/igc-admin-backend/middleware"
	"github.com/Mastermind730/igc-admin-backend/models"
	"github.com/Mastermind730/igc-admin-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup MongoDB connection
	client, err := SetupMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	
	// Create a database service
	dbService := models.NewDatabaseService(client, "pccoe_IGC")
	defer dbService.Close()
	
	fmt.Println("Connected to MongoDB successfully!")
	
	// Initialize handlers
	userHandler := handlers.NewUserHandler(dbService)
	teamHandler := handlers.NewTeamRegistrationHandler(dbService)
	
	// Create Gin router
	router := gin.New()
	
	// Add middleware
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length", 
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"Cache-Control",
		},
		AllowCredentials: true,
		MaxAge: 12 * 60 * 60, // 12 hours
	}

	router.Use(cors.New(corsConfig))
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(gin.Recovery())
	
	// Setup routes
	routes.SetupRoutes(router, userHandler, teamHandler)
	
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("🚀 IGC Admin Backend API Server starting on port %s\n", port)
	fmt.Printf("📖 API Documentation available at: http://localhost:%s/api/v1/health\n", port)
	fmt.Printf("🌐 Base URL: http://localhost:%s\n", port)
	
	// Print available routes
	fmt.Println("\n📋 Available API Routes:")
	fmt.Println("================================")
	fmt.Println("Authentication:")
	fmt.Println("  POST /api/v1/auth/login")
	fmt.Println("\nUser Management (Admin):")
	fmt.Println("  POST /api/v1/users")
	fmt.Println("  GET  /api/v1/users")
	fmt.Println("  GET  /api/v1/users/{id}")
	fmt.Println("  PUT  /api/v1/users/{id}")
	fmt.Println("  DELETE /api/v1/users/{id}")
	fmt.Println("\nTeam Registrations:")
	fmt.Println("  POST /api/v1/team-registrations")
	fmt.Println("  GET  /api/v1/team-registrations")
	fmt.Println("  GET  /api/v1/team-registrations/stats")
	fmt.Println("  GET  /api/v1/team-registrations/{id}")
	fmt.Println("  PUT  /api/v1/team-registrations/{id}")
	fmt.Println("  DELETE /api/v1/team-registrations/{id}")
	fmt.Println("  PUT  /api/v1/team-registrations/{id}/action")
	fmt.Println("  GET  /api/v1/team-registrations/reg/{regNumber}")
	fmt.Println("  GET  /api/v1/team-registrations/track/{track}")
	fmt.Println("\nHealth Check:")
	fmt.Println("  GET  /")
	fmt.Println("  GET  /api/v1/health")
	fmt.Println("================================")
	
	// Create a default admin user if none exists
	go createDefaultAdminUser(dbService)
	
	// Start the server
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// createDefaultAdminUser creates a default admin user if no users exist
func createDefaultAdminUser(db *models.DatabaseService) {
	count, err := db.CountUsers()
	if err != nil {
		log.Printf("Error checking user count: %v", err)
		return
	}
	
	if count == 0 {
		defaultUser := models.NewUser("admin", "admin123")
		createdUser, err := db.CreateUser(defaultUser)
		if err != nil {
			log.Printf("Error creating default admin user: %v", err)
			return
		}
		
		fmt.Printf("\n🔐 Default admin user created successfully!\n")
		fmt.Printf("   Username: %s\n", createdUser.Username)
		fmt.Printf("   Password: admin123\n")
		fmt.Printf("   ⚠️  Please change the default password after first login!\n\n")
	}
}
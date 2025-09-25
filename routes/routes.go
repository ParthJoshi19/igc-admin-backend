package routes

import (
	"github.com/Mastermind730/igc-admin-backend/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, userHandler *handlers.UserHandler, teamHandler *handlers.TeamRegistrationHandler) {
	// API version 1
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
		}

		// User routes (admin only - should be protected with middleware)
		users := api.Group("/users")
		users.Use(handlers.JWTAuthMiddleware())
		{
			users.POST("/", userHandler.CreateUser)           // Create new admin user
			users.GET("/", userHandler.GetAllUsers)           // Get all users with pagination
			users.GET("/:id", userHandler.GetUser)            // Get user by ID
			users.PUT("/:id", userHandler.UpdateUser)         // Update user
			users.DELETE("/:id", userHandler.DeleteUser)      // Delete user
		}

		// Team registration routes
		teams := api.Group("/team-registrations")
		teams.Use(handlers.JWTAuthMiddleware())
		{
			teams.POST("/", teamHandler.CreateTeamRegistration)              // Create new team registration
			teams.GET("/", teamHandler.GetAllTeamRegistrations)              // Get all teams with filters
			teams.GET("/stats", teamHandler.GetTeamRegistrationStats)        // Get registration statistics
			teams.GET("/:id", teamHandler.GetTeamRegistration)               // Get team by ID
			teams.PUT("/:id", teamHandler.UpdateTeamRegistration)            // Update team registration
			teams.DELETE("/:id", teamHandler.DeleteTeamRegistration)         // Delete team registration (admin)
			teams.PUT("/:id/action", teamHandler.ApproveOrRejectTeamRegistration) // Approve/Reject team (admin)
			teams.GET("/reg/:regNumber", teamHandler.GetTeamRegistrationByRegNumber) // Get team by registration number
			teams.GET("/track/:track", teamHandler.GetTeamRegistrationsByTrack)      // Get teams by track
			// New routes for allocation and judge evaluation
			teams.PUT("/:id/allocate", userHandler.AllocateTeamToJudge) // Admin allocates team to judge
			teams.GET("/allocated", userHandler.GetAllocatedTeamsForJudge) // Judge views allocated teams
			teams.PUT("/:id/evaluate", userHandler.JudgeEvaluateTeam) // Judge approves/rejects team
		}

		// Health check route
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"message": "IGC Admin Backend API is running",
				"version": "1.0.0",
			})
		})
	}
	router.POST("/api/v1/create-default-admin", userHandler.CreateDefaultAdmin)
	// Root health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to IGC Admin Backend API",
			"version": "1.0.0",
			"docs":    "/api/v1/health",
		})
	})
}
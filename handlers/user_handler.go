package handlers

import (
	"net/http"
	"strings"

	"github.com/Mastermind730/igc-admin-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	DB *models.DatabaseService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(db *models.DatabaseService) *UserHandler {
	return &UserHandler{DB: db}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest represents the create user request payload
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest represents the update user request payload
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Password string `json:"password,omitempty" binding:"omitempty,min=6"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param loginData body LoginRequest true "Login credentials"
// @Success 200 {object} UserResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get user by username
	user, err := h.DB.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// In a real application, you would hash and compare passwords
	// For now, we'll do a simple comparison (NOT SECURE - implement proper hashing)
	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Return user data (without password)
	response := UserResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    response,
	})
}

// CreateUser creates a new user (admin only)
// @Summary Create a new user
// @Description Create a new admin user
// @Tags users
// @Accept json
// @Produce json
// @Param userData body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} gin.H
// @Failure 409 {object} gin.H
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, _ := h.DB.GetUserByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Create new user
	newUser := models.NewUser(req.Username, req.Password)
	createdUser, err := h.DB.CreateUser(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	// Return user data (without password)
	response := UserResponse{
		ID:       createdUser.ID.Hex(),
		Username: createdUser.Username,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    response,
	})
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.DB.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID", "details": err.Error()})
		}
		return
	}

	// Return user data (without password)
	response := UserResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}

	c.JSON(http.StatusOK, gin.H{
		"user": response,
	})
}

// GetAllUsers retrieves all users with pagination
// @Summary Get all users
// @Description Get all users with optional pagination
// @Tags users
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {array} UserResponse
// @Router /api/users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse query parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p := parseInt(pageStr); p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l := parseInt(limitStr); l > 0 && l <= 100 {
			limit = l
		}
	}

	skip := int64((page - 1) * limit)
	users, err := h.DB.GetAllUsers(int64(limit), skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users", "details": err.Error()})
		return
	}

	// Convert to response format (without passwords)
	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:       user.ID.Hex(),
			Username: user.Username,
		})
	}

	// Get total count
	total, err := h.DB.CountUsers()
	if err != nil {
		total = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateUser updates an existing user
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param userData body UpdateUserRequest true "Updated user data"
// @Success 200 {object} UserResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Check if user exists
	existingUser, err := h.DB.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID", "details": err.Error()})
		}
		return
	}

	// Prepare update data
	updateData := bson.M{}
	if req.Username != "" && req.Username != existingUser.Username {
		// Check if new username already exists
		if existingUserWithUsername, _ := h.DB.GetUserByUsername(req.Username); existingUserWithUsername != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		updateData["username"] = req.Username
	}
	if req.Password != "" {
		updateData["password"] = req.Password
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	updatedUser, err := h.DB.UpdateUser(userID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
		return
	}

	// Return updated user data (without password)
	response := UserResponse{
		ID:       updatedUser.ID.Hex(),
		Username: updatedUser.Username,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    response,
	})
}

// DeleteUser deletes a user by ID
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Param id path string true "User ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	err := h.DB.DeleteUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// Helper function to parse integer from string
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	
	result := 0
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0
		}
		result = result*10 + int(char-'0')
	}
	return result
}
package handlers

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Mastermind730/igc-admin-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// UnifiedCreateUserRequest represents the unified create user request payload
// If role is judge, require name, email, organization
// If role is admin, require username and password
// Role must be either "admin" or "judge"
type UnifiedCreateUserRequest struct {
	Username     string `json:"username" binding:"required,min=3,max=50"` // for admin, also used as email for judge
	Password     string `json:"password" binding:"required,min=6"`         // for admin, judge can use judgeID as password
	Role         string `json:"role" binding:"required,oneof=admin judge"`
	Name         string `json:"name,omitempty"`         // for judge
	Organization string `json:"organization,omitempty"` // for judge
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

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey" // fallback for demo
	}
	return secret
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID.Hex(),
		"username":  user.Username,
		"role":      user.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24h expiry
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// JWTAuthMiddleware validates JWT token and sets user info in context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
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

	// Generate JWT token
	token, err := GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return user data (without password) and token
	response := UserResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    response,
		"token":   token,
	})
}

// CreateUser creates a new user (admin only)
// @Summary Create a new user
// @Description Create a new admin user
// @Tags users
// @Accept json
// @Produce json
// @Param userData body UnifiedCreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} gin.H
// @Failure 409 {object} gin.H
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req UnifiedCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, _ := h.DB.GetUserByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// For judge, generate judgeID and use as password if not provided
	judgeID := ""
	if req.Role == "judge" {
		if req.Name == "" || req.Organization == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Judge must have name and organization"})
			return
		}
		judgeID = "JUDGE-" + generateRandomID()
		if req.Password == "" {
			req.Password = judgeID // set password to judgeID if not provided
		}
	}

	// Create new user
	newUser := models.NewUser(req.Username, req.Password)
	newUser.Role = req.Role
	// Optionally, extend User model to store Name, Organization, JudgeID
	createdUser, err := h.DB.CreateUser(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	response := gin.H{
		"id":       createdUser.ID.Hex(),
		"username": req.Username,
		"role":     req.Role,
	}
	if req.Role == "judge" {
		response["judgeId"] = judgeID
		response["name"] = req.Name
		response["organization"] = req.Organization
		response["password"] = req.Password // for demo
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

// CreateDefaultAdmin creates a default admin user
func (h *UserHandler) CreateDefaultAdmin(c *gin.Context) {
    username := "admin"
    password := "igc#407@"

    // Check if admin already exists
    existingUser, _ := h.DB.GetUserByUsername(username)
    if existingUser != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Admin user already exists"})
        return
    }

    // Create new admin user
    newUser := models.NewUser(username, password)
    newUser.Role = "admin"
    createdUser, err := h.DB.CreateUser(newUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user", "details": err.Error()})
        return
    }

    response := UserResponse{
        ID:       createdUser.ID.Hex(),
        Username: createdUser.Username,
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Admin user created successfully",
        "user":    response,
    })
}

// CreateJudgeRequest represents the create judge request payload
type CreateJudgeRequest struct {
    Name         string `json:"name" binding:"required,min=3,max=100"`
    Email        string `json:"email" binding:"required,email"`
    Organization string `json:"organization" binding:"required,min=2,max=100"`
}

// CreateJudge creates a new judge user
func (h *UserHandler) CreateJudge(c *gin.Context) {
    var req CreateJudgeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
        return
    }

    // Check if judge already exists by email
    existingUser, _ := h.DB.GetUserByUsername(req.Email)
    if existingUser != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Judge with this email already exists"})
        return
    }

    // Generate unique judge ID
    judgeID := "JUDGE-" + generateRandomID()

    // Create new judge user
    newUser := models.NewUser(req.Email, judgeID) // password is judgeID for now
    newUser.Role = "judge"
    // Add extra fields to user model if needed (Name, Organization)
    // For now, store in Username and add judgeID to a custom field if you extend the model

    createdUser, err := h.DB.CreateUser(newUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create judge", "details": err.Error()})
        return
    }

    response := gin.H{
        "id":       createdUser.ID.Hex(),
        "judgeId":  judgeID,
        "name":     req.Name,
        "email":    req.Email,
        "organization": req.Organization,
        "role":     "judge",
        "password": judgeID, // for demo, return password as judgeID
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Judge created successfully",
        "judge":   response,
    })
}

// TeamAllocationRequest for admin to allocate team to judge
// judgeId is the ObjectID hex string of the judge user
// teamId is the ObjectID hex string of the team
// Only admin can call this
// Route: PUT /api/v1/team-registrations/:id/allocate
// Body: { "judgeId": "..." }
func (h *UserHandler) AllocateTeamToJudge(c *gin.Context) {
    teamId := c.Param("id")
    var req struct {
        JudgeId string `json:"judgeId" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
        return
    }
    // Only admin can allocate
    role, _ := c.Get("role")
    if role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can allocate teams"})
        return
    }
    // Update team with allocated judge
    update := bson.M{"allocatedJudgeId": req.JudgeId}
    updatedTeam, err := h.DB.UpdateTeamRegistration(teamId, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to allocate team", "details": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Team allocated to judge", "team": updatedTeam})
}

// Judge can view teams allocated to them
// Route: GET /api/v1/team-registrations/allocated
func (h *UserHandler) GetAllocatedTeamsForJudge(c *gin.Context) {
    role, _ := c.Get("role")
    userId, _ := c.Get("user_id")
    if role != "judge" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only judges can view allocated teams"})
        return
    }
    filter := bson.M{"allocatedJudgeId": userId}
    teams, err := h.DB.GetAllTeamRegistrations(100, 0, filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get allocated teams", "details": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// Judge can approve or reject allocated team
// Route: PUT /api/v1/team-registrations/:id/evaluate
// Body: { "decision": "approve"|"reject", "reason": "..." }
func (h *UserHandler) JudgeEvaluateTeam(c *gin.Context) {
    teamId := c.Param("id")
    role, _ := c.Get("role")
    userId, _ := c.Get("user_id")
    if role != "judge" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only judges can evaluate teams"})
        return
    }
    var req struct {
        Decision string `json:"decision" binding:"required,oneof=approve reject"`
        Reason   string `json:"reason"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
        return
    }
    update := bson.M{"actionedBy": userId}
    if req.Decision == "approve" {
        update["registrationStatus"] = "approved"
        update["approvedAt"] = time.Now()
    } else {
        update["registrationStatus"] = "rejected"
        update["rejectedAt"] = time.Now()
        update["rejectionReason"] = req.Reason
    }
    updatedTeam, err := h.DB.UpdateTeamRegistration(teamId, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team status", "details": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Team evaluation updated", "team": updatedTeam})
}

// generateRandomID generates a random string for judge ID
func generateRandomID() string {
    // Simple random string generator (for demo)
    rand.Seed(time.Now().UnixNano())
    letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    b := make([]rune, 6)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
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
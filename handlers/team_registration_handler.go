package handlers

import (
	"net/http"
	"strings"

	"github.com/Mastermind730/igc-admin-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// TeamRegistrationHandler handles team registration API requests
type TeamRegistrationHandler struct {
	DB *models.DatabaseService
}

// NewTeamRegistrationHandler creates a new TeamRegistrationHandler
func NewTeamRegistrationHandler(db *models.DatabaseService) *TeamRegistrationHandler {
	return &TeamRegistrationHandler{DB: db}
}

// CreateTeamRegistrationRequest represents the team registration request payload
type CreateTeamRegistrationRequest struct {
	TeamName              string                      `json:"teamName" binding:"required,max=100"`
	LeaderName            string                      `json:"leaderName" binding:"required,max=100"`
	LeaderEmail           string                      `json:"leaderEmail" binding:"required,email"`
	LeaderMobile          string                      `json:"leaderMobile" binding:"required"`
	LeaderGender          models.Gender               `json:"leaderGender" binding:"required"`
	Institution           string                      `json:"institution" binding:"required,max=200"`
	Program               models.Program              `json:"program" binding:"required"`
	Country               string                      `json:"country" binding:"required,max=100"`
	State                 string                      `json:"state" binding:"required,max=100"`
	Members               []models.TeamMember         `json:"members"`
	MentorName            string                      `json:"mentorName" binding:"required,max=100"`
	MentorEmail           string                      `json:"mentorEmail" binding:"required,email"`
	MentorMobile          string                      `json:"mentorMobile" binding:"required"`
	MentorInstitution     string                      `json:"mentorInstitution" binding:"required,max=200"`
	MentorDesignation     string                      `json:"mentorDesignation" binding:"required,max=100"`
	InstituteNOC          *models.DriveFile           `json:"instituteNOC,omitempty"`
	IDCardsPDF            *models.DriveFile           `json:"idCardsPDF,omitempty"`
	TopicName             string                      `json:"topicName" binding:"required,max=200"`
	TopicDescription      string                      `json:"topicDescription" binding:"required"`
	Track                 models.Track                `json:"track" binding:"required"`
	PresentationPPT       models.DriveFile            `json:"presentationPPT" binding:"required"`
}

// UpdateTeamRegistrationRequest represents the update team registration request payload
type UpdateTeamRegistrationRequest struct {
	TeamName              string                      `json:"teamName,omitempty" binding:"omitempty,max=100"`
	LeaderName            string                      `json:"leaderName,omitempty" binding:"omitempty,max=100"`
	LeaderEmail           string                      `json:"leaderEmail,omitempty" binding:"omitempty,email"`
	LeaderMobile          string                      `json:"leaderMobile,omitempty"`
	LeaderGender          *models.Gender              `json:"leaderGender,omitempty"`
	Institution           string                      `json:"institution,omitempty" binding:"omitempty,max=200"`
	Program               *models.Program             `json:"program,omitempty"`
	Country               string                      `json:"country,omitempty" binding:"omitempty,max=100"`
	State                 string                      `json:"state,omitempty" binding:"omitempty,max=100"`
	Members               []models.TeamMember         `json:"members,omitempty"`
	MentorName            string                      `json:"mentorName,omitempty" binding:"omitempty,max=100"`
	MentorEmail           string                      `json:"mentorEmail,omitempty" binding:"omitempty,email"`
	MentorMobile          string                      `json:"mentorMobile,omitempty"`
	MentorInstitution     string                      `json:"mentorInstitution,omitempty" binding:"omitempty,max=200"`
	MentorDesignation     string                      `json:"mentorDesignation,omitempty" binding:"omitempty,max=100"`
	InstituteNOC          *models.DriveFile           `json:"instituteNOC,omitempty"`
	IDCardsPDF            *models.DriveFile           `json:"idCardsPDF,omitempty"`
	TopicName             string                      `json:"topicName,omitempty" binding:"omitempty,max=200"`
	TopicDescription      string                      `json:"topicDescription,omitempty"`
	Track                 *models.Track               `json:"track,omitempty"`
	PresentationPPT       *models.DriveFile           `json:"presentationPPT,omitempty"`
}

// ApproveRejectRequest represents the approve/reject request payload
type ApproveRejectRequest struct {
	Action    string `json:"action" binding:"required,oneof=approve reject"`
	Reason    string `json:"reason,omitempty"`
	ActionedBy string `json:"actionedBy" binding:"required"`
}

// CreateTeamRegistration creates a new team registration
// @Summary Create team registration
// @Description Create a new team registration for the IGC hackathon
// @Tags team-registrations
// @Accept json
// @Produce json
// @Param teamData body CreateTeamRegistrationRequest true "Team registration data"
// @Success 201 {object} models.TeamRegistration
// @Failure 400 {object} gin.H
// @Failure 409 {object} gin.H
// @Router /api/team-registrations [post]
func (h *TeamRegistrationHandler) CreateTeamRegistration(c *gin.Context) {
	var req CreateTeamRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Validate team size (1-4 members + leader)
	validMembers := 0
	for _, member := range req.Members {
		if len(strings.TrimSpace(member.FullName)) > 0 {
			validMembers++
		}
	}
	if validMembers < 1 || validMembers > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team must have between 1-4 members (excluding leader)"})
		return
	}

	// Check if team name already exists
	existingTeam, _ := h.DB.GetTeamRegistrationByTeamName(req.TeamName)
	if existingTeam != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Team name already exists"})
		return
	}

	// Create new team registration
	teamReg := models.NewTeamRegistration()
	teamReg.TeamName = req.TeamName
	teamReg.LeaderName = req.LeaderName
	teamReg.LeaderEmail = req.LeaderEmail
	teamReg.LeaderMobile = req.LeaderMobile
	teamReg.LeaderGender = req.LeaderGender
	teamReg.Institution = req.Institution
	teamReg.Program = req.Program
	teamReg.Country = req.Country
	teamReg.State = req.State
	teamReg.Members = req.Members
	teamReg.MentorName = req.MentorName
	teamReg.MentorEmail = req.MentorEmail
	teamReg.MentorMobile = req.MentorMobile
	teamReg.MentorInstitution = req.MentorInstitution
	teamReg.MentorDesignation = req.MentorDesignation
	teamReg.InstituteNOC = req.InstituteNOC
	teamReg.IDCardsPDF = req.IDCardsPDF
	teamReg.TopicName = req.TopicName
	teamReg.TopicDescription = req.TopicDescription
	teamReg.Track = req.Track
	teamReg.PresentationPPT = req.PresentationPPT

	createdTeam, err := h.DB.CreateTeamRegistration(teamReg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team registration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team registration created successfully",
		"team":    createdTeam,
	})
}

// GetTeamRegistration retrieves a team registration by ID
// @Summary Get team registration by ID
// @Description Get team registration information by ID
// @Tags team-registrations
// @Produce json
// @Param id path string true "Team Registration ID"
// @Success 200 {object} models.TeamRegistration
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/team-registrations/{id} [get]
func (h *TeamRegistrationHandler) GetTeamRegistration(c *gin.Context) {
	teamID := c.Param("id")

	team, err := h.DB.GetTeamRegistrationByID(teamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team registration not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team registration ID", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team": team,
	})
}

// GetTeamRegistrationByRegNumber retrieves a team registration by registration number
// @Summary Get team registration by registration number
// @Description Get team registration information by registration number
// @Tags team-registrations
// @Produce json
// @Param regNumber path string true "Registration Number"
// @Success 200 {object} models.TeamRegistration
// @Failure 404 {object} gin.H
// @Router /api/team-registrations/reg/{regNumber} [get]
func (h *TeamRegistrationHandler) GetTeamRegistrationByRegNumber(c *gin.Context) {
	regNumber := c.Param("regNumber")

	team, err := h.DB.GetTeamRegistrationByRegistrationNumber(regNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team registration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team": team,
	})
}

// GetAllTeamRegistrations retrieves all team registrations with pagination and filtering
// @Summary Get all team registrations
// @Description Get all team registrations with optional pagination and filtering
// @Tags team-registrations
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Param status query string false "Filter by status (pending/approved/rejected)"
// @Param track query string false "Filter by track"
// @Param institution query string false "Filter by institution"
// @Success 200 {array} models.TeamRegistration
// @Router /api/team-registrations [get]
func (h *TeamRegistrationHandler) GetAllTeamRegistrations(c *gin.Context) {
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

	// Build filter
	filter := bson.M{}
	
	if status := c.Query("status"); status != "" {
		filter["registrationStatus"] = status
	}
	
	if track := c.Query("track"); track != "" {
		filter["track"] = track
	}
	
	if institution := c.Query("institution"); institution != "" {
		filter["institution"] = bson.M{"$regex": institution, "$options": "i"}
	}

	skip := int64((page - 1) * limit)
	teams, err := h.DB.GetAllTeamRegistrations(int64(limit), skip, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve team registrations", "details": err.Error()})
		return
	}

	// Get total count
	total, err := h.DB.CountTeamRegistrations()
	if err != nil {
		total = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTeamRegistrationsByTrack retrieves teams by track
// @Summary Get team registrations by track
// @Description Get team registrations filtered by track
// @Tags team-registrations
// @Produce json
// @Param track path string true "Track name"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {array} models.TeamRegistration
// @Router /api/team-registrations/track/{track} [get]
func (h *TeamRegistrationHandler) GetTeamRegistrationsByTrack(c *gin.Context) {
	track := models.Track(c.Param("track"))
	
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
	teams, err := h.DB.GetTeamRegistrationsByTrack(track, int64(limit), skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve team registrations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
		"track": track,
	})
}

// UpdateTeamRegistration updates an existing team registration
// @Summary Update team registration
// @Description Update team registration information
// @Tags team-registrations
// @Accept json
// @Produce json
// @Param id path string true "Team Registration ID"
// @Param teamData body UpdateTeamRegistrationRequest true "Updated team data"
// @Success 200 {object} models.TeamRegistration
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/team-registrations/{id} [put]
func (h *TeamRegistrationHandler) UpdateTeamRegistration(c *gin.Context) {
	teamID := c.Param("id")
	
	var req UpdateTeamRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Check if team exists
	_, err := h.DB.GetTeamRegistrationByID(teamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team registration not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team registration ID", "details": err.Error()})
		}
		return
	}

	// Prepare update data
	updateData := bson.M{}
	if req.TeamName != "" {
		updateData["teamName"] = req.TeamName
	}
	if req.LeaderName != "" {
		updateData["leaderName"] = req.LeaderName
	}
	if req.LeaderEmail != "" {
		updateData["leaderEmail"] = req.LeaderEmail
	}
	if req.LeaderMobile != "" {
		updateData["leaderMobile"] = req.LeaderMobile
	}
	if req.LeaderGender != nil {
		updateData["leaderGender"] = *req.LeaderGender
	}
	if req.Institution != "" {
		updateData["institution"] = req.Institution
	}
	if req.Program != nil {
		updateData["program"] = *req.Program
	}
	if req.Country != "" {
		updateData["country"] = req.Country
	}
	if req.State != "" {
		updateData["state"] = req.State
	}
	if req.Members != nil {
		updateData["members"] = req.Members
	}
	if req.MentorName != "" {
		updateData["mentorName"] = req.MentorName
	}
	if req.MentorEmail != "" {
		updateData["mentorEmail"] = req.MentorEmail
	}
	if req.MentorMobile != "" {
		updateData["mentorMobile"] = req.MentorMobile
	}
	if req.MentorInstitution != "" {
		updateData["mentorInstitution"] = req.MentorInstitution
	}
	if req.MentorDesignation != "" {
		updateData["mentorDesignation"] = req.MentorDesignation
	}
	if req.InstituteNOC != nil {
		updateData["instituteNOC"] = req.InstituteNOC
	}
	if req.IDCardsPDF != nil {
		updateData["idCardsPDF"] = req.IDCardsPDF
	}
	if req.TopicName != "" {
		updateData["topicName"] = req.TopicName
	}
	if req.TopicDescription != "" {
		updateData["topicDescription"] = req.TopicDescription
	}
	if req.Track != nil {
		updateData["track"] = *req.Track
	}
	if req.PresentationPPT != nil {
		updateData["presentationPPT"] = req.PresentationPPT
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	updatedTeam, err := h.DB.UpdateTeamRegistration(teamID, updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team registration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team registration updated successfully",
		"team":    updatedTeam,
	})
}

// ApproveOrRejectTeamRegistration approves or rejects a team registration
// @Summary Approve or reject team registration
// @Description Approve or reject a team registration (admin only)
// @Tags team-registrations
// @Accept json
// @Produce json
// @Param id path string true "Team Registration ID"
// @Param actionData body ApproveRejectRequest true "Action data"
// @Success 200 {object} models.TeamRegistration
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/team-registrations/{id}/action [put]
func (h *TeamRegistrationHandler) ApproveOrRejectTeamRegistration(c *gin.Context) {
	teamID := c.Param("id")
	
	var req ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	var updatedTeam *models.TeamRegistration
	var err error

	if req.Action == "approve" {
		updatedTeam, err = h.DB.ApproveTeamRegistration(teamID, req.ActionedBy)
	} else if req.Action == "reject" {
		if req.Reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required"})
			return
		}
		updatedTeam, err = h.DB.RejectTeamRegistration(teamID, req.Reason, req.ActionedBy)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team registration not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team registration", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team registration " + req.Action + "d successfully",
		"team":    updatedTeam,
	})
}

// DeleteTeamRegistration deletes a team registration by ID
// @Summary Delete team registration
// @Description Delete a team registration by ID (admin only)
// @Tags team-registrations
// @Param id path string true "Team Registration ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /api/team-registrations/{id} [delete]
func (h *TeamRegistrationHandler) DeleteTeamRegistration(c *gin.Context) {
	teamID := c.Param("id")

	err := h.DB.DeleteTeamRegistration(teamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team registration not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team registration ID", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team registration deleted successfully",
	})
}

// GetTeamRegistrationStats retrieves registration statistics
// @Summary Get team registration statistics
// @Description Get statistics about team registrations
// @Tags team-registrations
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/team-registrations/stats [get]
func (h *TeamRegistrationHandler) GetTeamRegistrationStats(c *gin.Context) {
	stats, err := h.DB.GetTeamRegistrationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
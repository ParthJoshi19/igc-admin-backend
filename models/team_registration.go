package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Gender enum type for validation
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// Program enum type for validation
type Program string

const (
	ProgramBTechCS     Program = "B.Tech - Computer Engineering"
	ProgramBTechIT     Program = "B.Tech - Information Technology"
	ProgramBTechEC     Program = "B.Tech - Electronics & Telecommunication"
	ProgramBTechMech   Program = "B.Tech - Mechanical Engineering"
	ProgramBTechCivil  Program = "B.Tech - Civil Engineering"
	ProgramBTechEE     Program = "B.Tech - Electrical Engineering"
	ProgramMTechCS     Program = "M.Tech - Computer Engineering"
	ProgramMTechIT     Program = "M.Tech - Information Technology"
	ProgramMTechEC     Program = "M.Tech - Electronics & Telecommunication"
	ProgramMCA         Program = "MCA - Master of Computer Applications"
	ProgramMBA         Program = "MBA - Master of Business Administration"
	ProgramOther       Program = "Other"
)

// Track enum type for validation
type Track string

const (
	TrackClimateForecasting      Track = "Climate Forecasting"
	TrackSmartAgriculture        Track = "Smart Agriculture"
	TrackDisasterManagement      Track = "Disaster Management"
	TrackGreenTransportation     Track = "Green Transportation"
	TrackEnergyOptimization      Track = "Energy Optimization"
	TrackWaterConservation       Track = "Water Conservation"
	TrackCarbonTracking          Track = "Carbon Tracking"
	TrackBiodiversityMonitoring  Track = "Biodiversity Monitoring"
	TrackSustainableCities       Track = "Sustainable Cities"
	TrackWasteManagement         Track = "Waste Management"
	TrackAirQuality              Track = "Air Quality"
	TrackDeforestationPrevention Track = "Deforestation Prevention"
	TrackClimateEducation        Track = "Climate Education"
	TrackAIEnvironmentalData     Track = "AI-based Environmental Data Analysis"
	TrackPublicHealthClimate     Track = "Public Health Impact of Climate Change"
	TrackOceanMarine             Track = "Ocean & Marine Protection using AI"
)

// RegistrationStatus enum type
type RegistrationStatus string

const (
	StatusPending  RegistrationStatus = "pending"
	StatusApproved RegistrationStatus = "approved"
	StatusRejected RegistrationStatus = "rejected"
)

// TeamMember represents a team member (excluding leader)
type TeamMember struct {
	FullName string `bson:"fullName" json:"fullName" validate:"required,max=100"`
	Gender   Gender `bson:"gender" json:"gender" validate:"required,oneof=male female other"`
	MobileNo string `bson:"mobileNo" json:"mobileNo" validate:"required,e164"`
	Email    string `bson:"email" json:"email" validate:"required,email,lowercase"`
}

// DriveFile represents a Cloudinary file URL
type DriveFile struct {
	FileURL string `bson:"fileUrl" json:"fileUrl" validate:"required,url"`
}

// TeamRegistration represents the complete team registration
type TeamRegistration struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	TeamName         string              `bson:"teamName" json:"teamName" validate:"required,max=100"`
	LeaderName       string              `bson:"leaderName" json:"leaderName" validate:"required,max=100"`
	LeaderEmail      string              `bson:"leaderEmail" json:"leaderEmail" validate:"required,email,lowercase"`
	LeaderMobile     string              `bson:"leaderMobile" json:"leaderMobile" validate:"required,e164"`
	LeaderGender     Gender              `bson:"leaderGender" json:"leaderGender" validate:"required,oneof=male female other"`
	Institution      string              `bson:"institution" json:"institution" validate:"required,max=200"`
	Program          Program             `bson:"program" json:"program" validate:"required"`
	Country          string              `bson:"country" json:"country" validate:"required,max=100"`
	State            string              `bson:"state" json:"state" validate:"required,max=100"`
	Members          []TeamMember        `bson:"members" json:"members" validate:"dive,min=1,max=4"`
	MentorName       string              `bson:"mentorName" json:"mentorName" validate:"required,max=100"`
	MentorEmail      string              `bson:"mentorEmail" json:"mentorEmail" validate:"required,email,lowercase"`
	MentorMobile     string              `bson:"mentorMobile" json:"mentorMobile" validate:"required,e164"`
	MentorInstitution string             `bson:"mentorInstitution" json:"mentorInstitution" validate:"required,max=200"`
	MentorDesignation string             `bson:"mentorDesignation" json:"mentorDesignation" validate:"required,max=100"`
	InstituteNOC     *DriveFile          `bson:"instituteNOC,omitempty" json:"instituteNOC,omitempty"`
	IDCardsPDF       *DriveFile          `bson:"idCardsPDF,omitempty" json:"idCardsPDF,omitempty"`
	TopicName        string              `bson:"topicName" json:"topicName" validate:"required,max=200"`
	TopicDescription string              `bson:"topicDescription" json:"topicDescription" validate:"required"`
	Track            Track               `bson:"track" json:"track" validate:"required"`
	PresentationPPT  DriveFile           `bson:"presentationPPT" json:"presentationPPT" validate:"required"`
	
	// Status and tracking fields
	RegistrationStatus RegistrationStatus `bson:"registrationStatus" json:"registrationStatus"`
	RegistrationNumber string             `bson:"registrationNumber,omitempty" json:"registrationNumber,omitempty"`
	TeamID             string             `bson:"teamId,omitempty" json:"teamId,omitempty"`
	
	// Timestamps
	SubmittedAt time.Time  `bson:"submittedAt" json:"submittedAt"`
	ApprovedAt  *time.Time `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`
	RejectedAt  *time.Time `bson:"rejectedAt,omitempty" json:"rejectedAt,omitempty"`
	CreatedAt   time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `bson:"updatedAt" json:"updatedAt"`
	
	// Action tracking
	RejectionReason string `bson:"rejectionReason,omitempty" json:"rejectionReason,omitempty" validate:"max=500"`
	ActionedBy      string `bson:"actionedBy,omitempty" json:"actionedBy,omitempty" validate:"max=100"`
    AllocatedJudgeID primitive.ObjectID `bson:"allocatedJudgeId,omitempty" json:"allocatedJudgeId,omitempty"`
}

// NewTeamRegistration creates a new team registration with default values
func NewTeamRegistration() *TeamRegistration {
	now := time.Now()
	return &TeamRegistration{
		RegistrationStatus: StatusPending,
		SubmittedAt:        now,
		CreatedAt:          now,
		UpdatedAt:          now,
		Members:            make([]TeamMember, 0),
	}
}

// GetTeamSize returns the total team size (including leader)
func (tr *TeamRegistration) GetTeamSize() int {
	validMembers := 0
	for _, member := range tr.Members {
		if len(member.FullName) > 0 {
			validMembers++
		}
	}
	return validMembers + 1 // +1 for leader
}

// Approve marks the team registration as approved
func (tr *TeamRegistration) Approve(actionedBy string) {
	now := time.Now()
	tr.RegistrationStatus = StatusApproved
	tr.ApprovedAt = &now
	tr.ActionedBy = actionedBy
	tr.UpdatedAt = now
}

// Reject marks the team registration as rejected
func (tr *TeamRegistration) Reject(reason, actionedBy string) {
	now := time.Now()
	tr.RegistrationStatus = StatusRejected
	tr.RejectionReason = reason
	tr.RejectedAt = &now
	tr.ActionedBy = actionedBy
	tr.UpdatedAt = now
}

// UpdateTimestamp updates the UpdatedAt field to current time
func (tr *TeamRegistration) UpdateTimestamp() {
	tr.UpdatedAt = time.Now()
}

// IsApproved checks if the team registration is approved
func (tr *TeamRegistration) IsApproved() bool {
	return tr.RegistrationStatus == StatusApproved
}

// IsRejected checks if the team registration is rejected
func (tr *TeamRegistration) IsRejected() bool {
	return tr.RegistrationStatus == StatusRejected
}

// IsPending checks if the team registration is pending
func (tr *TeamRegistration) IsPending() bool {
	return tr.RegistrationStatus == StatusPending
}
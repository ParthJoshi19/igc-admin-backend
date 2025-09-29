package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database service struct
type DatabaseService struct {
	Client         *mongo.Client
	Database       *mongo.Database
	UserCollection *mongo.Collection
	TeamCollection *mongo.Collection
	Videos         *mongo.Collection
}

// NewDatabaseService creates a new database service
func NewDatabaseService(client *mongo.Client, dbName string) *DatabaseService {
	db := client.Database(dbName)

	return &DatabaseService{
		Client:         client,
		Database:       db,
		UserCollection: db.Collection("users"),
		TeamCollection: db.Collection("teamregistrations"),
		Videos:         db.Collection("videos"),
	}
}

// getContext creates a new context with timeout for database operations
func (db *DatabaseService) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// User CRUD Operations

// CreateUser creates a new user in the database
func (db *DatabaseService) CreateUser(user *User) (*User, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := db.UserCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// GetUserByID retrieves a user by their ID
func (db *DatabaseService) GetUserByID(id string) (*User, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var user User
	filter := bson.M{"_id": objectID}
	err = db.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (db *DatabaseService) GetUserByUsername(username string) (*User, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	var user User
	filter := bson.M{"username": username}
	err := db.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database
func (db *DatabaseService) GetAllUsers(limit int64, skip int64) ([]*User, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	opts := options.Find().SetLimit(limit).SetSkip(skip)
	cursor, err := db.UserCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates an existing user
func (db *DatabaseService) UpdateUser(id string, updateData bson.M) (*User, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	updateData["updatedAt"] = time.Now()
	update := bson.M{"$set": updateData}
	filter := bson.M{"_id": objectID}

	_, err = db.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return db.GetUserByID(id)
}

// DeleteUser deletes a user by ID
func (db *DatabaseService) DeleteUser(id string) error {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	filter := bson.M{"_id": objectID}
	result, err := db.UserCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// CountUsers returns the total number of users
func (db *DatabaseService) CountUsers() (int64, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	count, err := db.UserCollection.CountDocuments(ctx, bson.M{})
	return count, err
}

// Team Registration CRUD Operations

// CreateTeamRegistration creates a new team registration
func (db *DatabaseService) CreateTeamRegistration(team *TeamRegistration) (*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	// Generate registration number and team ID
	count, err := db.TeamCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	team.ID = primitive.NewObjectID()
	team.RegistrationNumber = fmt.Sprintf("PCCOEIGC%03d", count+1)
	team.TeamID = fmt.Sprintf("IGC%03d", count+1)
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	team.SubmittedAt = time.Now()

	result, err := db.TeamCollection.InsertOne(ctx, team)
	if err != nil {
		return nil, err
	}

	team.ID = result.InsertedID.(primitive.ObjectID)
	return team, nil
}

// GetTeamRegistrationByID retrieves a team registration by ID
func (db *DatabaseService) GetTeamRegistrationByID(id string) (*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid team registration ID format")
	}

	var team TeamRegistration
	filter := bson.M{"_id": objectID}
	err = db.TeamCollection.FindOne(ctx, filter).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("team registration not found")
		}
		return nil, err
	}

	return &team, nil
}

// GetTeamRegistrationByTeamName retrieves a team registration by team name
func (db *DatabaseService) GetTeamRegistrationByTeamName(teamName string) (*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	var team TeamRegistration
	filter := bson.M{"teamName": teamName}
	err := db.TeamCollection.FindOne(ctx, filter).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("team registration not found")
		}
		return nil, err
	}

	return &team, nil
}

// GetTeamRegistrationByRegistrationNumber retrieves a team by registration number
func (db *DatabaseService) GetTeamRegistrationByRegistrationNumber(regNumber string) (*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	var team TeamRegistration
	filter := bson.M{"registrationNumber": regNumber}
	err := db.TeamCollection.FindOne(ctx, filter).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("team registration not found")
		}
		return nil, err
	}

	return &team, nil
}

// GetAllTeamRegistrations retrieves all team registrations with pagination and filtering
func (db *DatabaseService) GetAllTeamRegistrations(limit int64, skip int64, filter bson.M) ([]*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	opts := options.Find().SetLimit(limit).SetSkip(skip).SetSort(bson.M{"submittedAt": -1})
	cursor, err := db.TeamCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []*TeamRegistration
	for cursor.Next(ctx) {
		var team TeamRegistration
		if err := cursor.Decode(&team); err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// GetTeamRegistrationsByTrack retrieves teams by track
func (db *DatabaseService) GetTeamRegistrationsByTrack(track Track, limit int64, skip int64) ([]*TeamRegistration, error) {
	filter := bson.M{"track": track}
	return db.GetAllTeamRegistrations(limit, skip, filter)
}

// GetTeamRegistrationsByStatus retrieves teams by registration status
func (db *DatabaseService) GetTeamRegistrationsByStatus(status RegistrationStatus, limit int64, skip int64) ([]*TeamRegistration, error) {
	filter := bson.M{"registrationStatus": status}
	return db.GetAllTeamRegistrations(limit, skip, filter)
}

// GetTeamRegistrationsByInstitution retrieves teams by institution
func (db *DatabaseService) GetTeamRegistrationsByInstitution(institution string, limit int64, skip int64) ([]*TeamRegistration, error) {
	filter := bson.M{"institution": institution}
	return db.GetAllTeamRegistrations(limit, skip, filter)
}

// UpdateTeamRegistration updates an existing team registration
func (db *DatabaseService) UpdateTeamRegistration(id string, updateData bson.M) (*TeamRegistration, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid team registration ID format")
	}

	updateData["updatedAt"] = time.Now()
	update := bson.M{"$set": updateData}
	filter := bson.M{"_id": objectID}

	_, err = db.TeamCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return db.GetTeamRegistrationByID(id)
}

// ApproveTeamRegistration approves a team registration
func (db *DatabaseService) ApproveTeamRegistration(id, actionedBy string) (*TeamRegistration, error) {
	team, err := db.GetTeamRegistrationByID(id)
	if err != nil {
		return nil, err
	}

	team.Approve(actionedBy)

	updateData := bson.M{
		"registrationStatus": StatusApproved,
		"approvedAt":         team.ApprovedAt,
		"actionedBy":         actionedBy,
		"updatedAt":          time.Now(),
	}

	return db.UpdateTeamRegistration(id, updateData)
}

// RejectTeamRegistration rejects a team registration
func (db *DatabaseService) RejectTeamRegistration(id, reason, actionedBy string) (*TeamRegistration, error) {
	team, err := db.GetTeamRegistrationByID(id)
	if err != nil {
		return nil, err
	}

	team.Reject(reason, actionedBy)

	updateData := bson.M{
		"registrationStatus": StatusRejected,
		"rejectionReason":    reason,
		"rejectedAt":         team.RejectedAt,
		"actionedBy":         actionedBy,
		"updatedAt":          time.Now(),
	}

	return db.UpdateTeamRegistration(id, updateData)
}

// DeleteTeamRegistration deletes a team registration by ID
func (db *DatabaseService) DeleteTeamRegistration(id string) error {
	ctx, cancel := db.getContext()
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid team registration ID format")
	}

	filter := bson.M{"_id": objectID}
	result, err := db.TeamCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("team registration not found")
	}

	return nil
}

// CountTeamRegistrations returns the total number of team registrations
func (db *DatabaseService) CountTeamRegistrations() (int64, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	count, err := db.TeamCollection.CountDocuments(ctx, bson.M{})
	return count, err
}

// CountTeamRegistrationsWithFilter returns the number of team registrations matching a filter
func (db *DatabaseService) CountTeamRegistrationsWithFilter(filter bson.M) (int64, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	if filter == nil {
		filter = bson.M{}
	}
	count, err := db.TeamCollection.CountDocuments(ctx, filter)
	return count, err
}

// GetVideoLinkForTeam returns the submitted video link for a team if present.
// It looks up in the "videos" collection using common identifiers.
func (db *DatabaseService) GetVideoLinkForTeam(team *TeamRegistration) (string, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	if db.Videos == nil || team == nil {
		return "", nil
	}

	// Try matching by teamId, registrationNumber, or teamName
	filter := bson.M{"$or": []bson.M{
		{"teamId": team.TeamID},
		{"registrationNumber": team.RegistrationNumber},
		{"teamName": team.TeamName},
	}}

	var doc bson.M
	err := db.Videos.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	// Common field names where a URL might be stored
	candidates := []string{"videoUrl", "videoURL", "videoLink", "link", "url"}
	for _, k := range candidates {
		if v, ok := doc[k]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
	}
	return "", nil
}

// CountTeamRegistrationsByStatus returns count by status
func (db *DatabaseService) CountTeamRegistrationsByStatus(status RegistrationStatus) (int64, error) {
	ctx, cancel := db.getContext()
	defer cancel()

	filter := bson.M{"registrationStatus": status}
	count, err := db.TeamCollection.CountDocuments(ctx, filter)
	return count, err
}

// GetTeamRegistrationStats returns registration statistics
func (db *DatabaseService) GetTeamRegistrationStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count total registrations
	total, err := db.CountTeamRegistrations()
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by status
	approved, err := db.CountTeamRegistrationsByStatus(StatusApproved)
	if err != nil {
		return nil, err
	}
	stats["approved"] = approved

	pending, err := db.CountTeamRegistrationsByStatus(StatusPending)
	if err != nil {
		return nil, err
	}
	stats["pending"] = pending

	rejected, err := db.CountTeamRegistrationsByStatus(StatusRejected)
	if err != nil {
		return nil, err
	}
	stats["rejected"] = rejected

	return stats, nil
}

// Close closes the database connection
func (db *DatabaseService) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Client.Disconnect(ctx); err != nil {
		fmt.Printf("Error disconnecting from MongoDB: %v\n", err)
	}
}

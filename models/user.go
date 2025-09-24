package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the MongoDB database (Admin/Staff)
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string             `bson:"username" json:"username" validate:"required,min=3,max=50"`
	Password  string             `bson:"password" json:"password,omitempty" validate:"required,min=6"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// NewUser creates a new user with default values
func NewUser(username, password string) *User {
	return &User{
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// UpdateTimestamp updates the UpdatedAt field to current time
func (u *User) UpdateTimestamp() {
	u.UpdatedAt = time.Now()
}
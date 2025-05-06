package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username"`
	Email        string             `json:"email" bson:"email"`
	Organization string             `json:"organization" bson:"organization"`
	Password     string             `json:"password" bson:"password"`
	Token        string             `json:"token,omitempty" bson:"token,omitempty"`
	TokenExpiry  time.Time          `json:"token_expiry,omitempty" bson:"token_expiry,omitempty"`
	RefreshToken string             `json:"refresh_token,omitempty" bson:"refresh_token,omitempty"`
}

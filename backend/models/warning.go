package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Level 1: Normal
// Level 2: Warning
// Level 3: Critical
// Level 4: Emergency

type Warning struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Level int                `json:"level" bson:"level"`
	Name  string             `json:"name" bson:"name"`
}

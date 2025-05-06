package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SensorConfig struct {
	FMax     int `json:"fmax" bson:"fmax"`
	LOR      int `json:"lor" bson:"lor"`
	GMax     int `json:"g_max" bson:"g_max"`
	AlarmThs int `json:"alarm_ths" bson:"alarm_ths"`
}

type Sensor struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	SerialNumber string             `json:"serial_number" bson:"serial_number"`
	Location     string             `json:"location" bson:"location"`
	Picture      string             `json:"picture" bson:"picture"`
	Config       SensorConfig       `json:"config" bson:"config"`
	Token        string             `json:"token,omitempty" bson:"token,omitempty"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VibrationData struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SensorID  primitive.ObjectID `bson:"sensor_id" json:"sensor_id"`
	WarnID    primitive.ObjectID `bson:"warn_id" json:"warn_id"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`

	// Acceleration in g units
	X_Axisg float32 `bson:"x_axisg" json:"x_axisg"` // X-axis acceleration in g
	Y_Axisg float32 `bson:"y_axisg" json:"y_axisg"` // Y-axis acceleration in g
	Z_Axisg float32 `bson:"z_axisg" json:"z_axisg"` // Z-axis acceleration in g

	// Acceleration in mm/s² units
	X_Axismm_s2 float32 `bson:"x_axismm_s2" json:"x_axismm_s2"` // X-axis acceleration in mm/s²
	Y_Axismm_s2 float32 `bson:"y_axismm_s2" json:"y_axismm_s2"` // Y-axis acceleration in mm/s²
	Z_Axismm_s2 float32 `bson:"z_axismm_s2" json:"z_axismm_s2"` // Z-axis acceleration in mm/s²

	// Velocity in mm/s units
	X_Axismm_s float32 `bson:"x_axismm_s" json:"x_axismm_s"` // X-axis velocity in mm/s
	Y_Axismm_s float32 `bson:"y_axismm_s" json:"y_axismm_s"` // Y-axis velocity in mm/s
	Z_Axismm_s float32 `bson:"z_axismm_s" json:"z_axismm_s"` // Z-axis velocity in mm/s
}

package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateVibration(c *gin.Context) {
	var vibration models.VibrationData
	if err := c.ShouldBindJSON(&vibration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate sensor ID
	if vibration.SensorID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sensor ID is required"})
		return
	}

	// Check if sensor exists
	sensorCollection := config.GetCollection("sensors")
	var sensor models.Sensor
	err := sensorCollection.FindOne(context.Background(), bson.M{"_id": vibration.SensorID}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	// Validate warning ID if provided
	if !vibration.WarnID.IsZero() {
		warningCollection := config.GetCollection("warnings")
		var warning models.Warning
		err := warningCollection.FindOne(context.Background(), bson.M{"_id": vibration.WarnID}).Decode(&warning)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warning ID"})
			return
		}
	}

	if vibration.Timestamp.IsZero() {
		vibration.Timestamp = time.Now()
	}

	collection := config.GetCollection("vibrations")
	result, err := collection.InsertOne(context.Background(), vibration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}

	vibration.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, vibration)
}

func GetVibrations(c *gin.Context) {
	var vibrations []models.VibrationData
	collection := config.GetCollection("vibrations")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip := (page - 1) * limit

	filter := bson.M{}

	if sensorID := c.Query("sensor_id"); sensorID != "" {
		if id, err := primitive.ObjectIDFromHex(sensorID); err == nil {
			filter["sensor_id"] = id
		}
	}

	if warnID := c.Query("warn_id"); warnID != "" {
		if id, err := primitive.ObjectIDFromHex(warnID); err == nil {
			filter["warn_id"] = id
		}
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter["timestamp"] = bson.M{"$gte": t}
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			if _, ok := filter["timestamp"]; ok {
				filter["timestamp"].(bson.M)["$lte"] = t
			} else {
				filter["timestamp"] = bson.M{"$lte": t}
			}
		}
	}

	// Add pagination options
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var vib models.VibrationData
		cursor.Decode(&vib)
		vibrations = append(vibrations, vib)
	}

	// Get total count for pagination
	total, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": vibrations,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func GetVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var vib models.VibrationData
	collection := config.GetCollection("vibrations")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&vib)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, vib)
}

func UpdateVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var vib models.VibrationData
	if err := c.ShouldBindJSON(&vib); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("vibrations")
	update := bson.M{
		"$set": bson.M{
			"sensor_id":   vib.SensorID,
			"warn_id":     vib.WarnID,
			"timestamp":   vib.Timestamp,
			"x_axisg":     vib.X_Axisg,
			"y_axisg":     vib.Y_Axisg,
			"z_axisg":     vib.Z_Axisg,
			"x_axismm_s2": vib.X_Axismm_s2,
			"y_axismm_s2": vib.Y_Axismm_s2,
			"z_axismm_s2": vib.Z_Axismm_s2,
			"x_axismm_s":  vib.X_Axismm_s,
			"y_axismm_s":  vib.Y_Axismm_s,
			"z_axismm_s":  vib.Z_Axismm_s,
		},
	}

	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vibration data updated"})
}

func DeleteVibration(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := config.GetCollection("vibrations")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vibration data not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vibration data deleted"})
}

func BatchRegisterVibrations(c *gin.Context) {
	var vibrations []models.VibrationData
	if err := c.ShouldBindJSON(&vibrations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate each vibration entry
	for _, vibration := range vibrations {
		// Validate sensor ID
		if vibration.SensorID.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Sensor ID is required for all entries"})
			return
		}

		// Check if sensor exists
		sensorCollection := config.GetCollection("sensors")
		var sensor models.Sensor
		err := sensorCollection.FindOne(context.Background(), bson.M{"_id": vibration.SensorID}).Decode(&sensor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID: " + vibration.SensorID.Hex()})
			return
		}

		// Validate warning ID if provided
		if !vibration.WarnID.IsZero() {
			warningCollection := config.GetCollection("warnings")
			var warning models.Warning
			err := warningCollection.FindOne(context.Background(), bson.M{"_id": vibration.WarnID}).Decode(&warning)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warning ID: " + vibration.WarnID.Hex()})
				return
			}
		}

		// Set timestamp if not provided
		if vibration.Timestamp.IsZero() {
			vibration.Timestamp = time.Now()
		}
	}

	// Prepare documents for bulk insert
	documents := make([]interface{}, len(vibrations))
	for i, vibration := range vibrations {
		documents[i] = vibration
	}

	collection := config.GetCollection("vibrations")
	result, err := collection.InsertMany(context.Background(), documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Batch insert failed"})
		return
	}

	// Update IDs in the response
	for i, id := range result.InsertedIDs {
		vibrations[i].ID = id.(primitive.ObjectID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully registered batch of vibration data",
		"count":   len(vibrations),
		"data":    vibrations,
	})
}

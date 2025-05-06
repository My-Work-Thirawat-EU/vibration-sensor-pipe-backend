package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	result, err := collection.InsertOne(context.Background(), sensor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sensor.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, sensor)
}

func GetSensors(c *gin.Context) {
	var sensors []models.Sensor
	collection := config.GetCollection("sensors")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var sensor models.Sensor
		cursor.Decode(&sensor)
		sensors = append(sensors, sensor)
	}

	c.JSON(http.StatusOK, sensors)
}

func GetSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var sensor models.Sensor
	collection := config.GetCollection("sensors")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, sensor)
}

func UpdateSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	update := bson.M{
		"$set": bson.M{
			"user_id":       sensor.UserID,
			"serial_number": sensor.SerialNumber,
			"location":      sensor.Location,
			"picture":       sensor.Picture,
			"config": bson.M{
				"fmax":      sensor.Config.FMax,
				"lor":       sensor.Config.LOR,
				"g_max":     sensor.Config.GMax,
				"alarm_ths": sensor.Config.AlarmThs,
			},
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor updated successfully"})
}

func DeleteSensor(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := config.GetCollection("sensors")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

func generateTokenHex(length int) (string, error) {
	tokenBytes := make([]byte, length)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func RegisterSensor(c *gin.Context) {
	var request struct {
		SerialNumber string `json:"serial_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	var sensor models.Sensor

	// Find sensor by serial number
	err := collection.FindOne(context.Background(), bson.M{"serial_number": request.SerialNumber}).Decode(&sensor)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	// Generate 32 bytes token (will become 64 hex characters)
	tokenString, err := generateTokenHex(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Update sensor document with the generated token
	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": sensor.ID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating sensor token"})
		return
	}

	// Extra safety: check if the sensor matched
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found during update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"sensor_id": sensor.ID.Hex(),
	})
}

func BatchRegisterSensors(c *gin.Context) {
	var sensors []models.Sensor
	if err := c.ShouldBindJSON(&sensors); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("sensors")
	var results []models.Sensor
	var errors []string

	for _, sensor := range sensors {
		// Generate token for each sensor
		tokenString, err := generateTokenHex(32)
		if err != nil {
			errors = append(errors, "Error generating token for sensor: "+sensor.SerialNumber)
			continue
		}
		sensor.Token = tokenString

		result, err := collection.InsertOne(context.Background(), sensor)
		if err != nil {
			errors = append(errors, "Error creating sensor: "+sensor.SerialNumber)
			continue
		}

		sensor.ID = result.InsertedID.(primitive.ObjectID)
		results = append(results, sensor)
	}

	response := gin.H{
		"successful_registrations": len(results),
		"failed_registrations":     len(errors),
		"sensors":                  results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if len(results) > 0 {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

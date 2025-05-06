package controllers

import (
	"context"
	"net/http"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var defaultWarnings = []models.Warning{
	{Level: 1, Name: "Normal"},
	{Level: 2, Name: "Warning"},
	{Level: 3, Name: "Critical"},
	{Level: 4, Name: "Emergency"},
}

func InitializeWarnings() error {
	collection := config.GetCollection("warnings")

	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	if count == 0 {
		var documents []interface{}
		for _, warning := range defaultWarnings {
			documents = append(documents, warning)
		}

		_, err := collection.InsertMany(context.Background(), documents)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetWarnings(c *gin.Context) {
	var warnings []models.Warning
	collection := config.GetCollection("warnings")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var warning models.Warning
		cursor.Decode(&warning)
		warnings = append(warnings, warning)
	}

	c.JSON(http.StatusOK, warnings)
}

func GetWarning(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var warning models.Warning
	collection := config.GetCollection("warnings")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&warning)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warning not found"})
		return
	}

	c.JSON(http.StatusOK, warning)
}

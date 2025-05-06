package config

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() error {
	// Set client options

	clientOptions := options.Client().ApplyURI("mongodb+srv://ThirawatEu:Wv7E6WwMcDJSs08s@vibration-sensor.ca1256c.mongodb.net/?retryWrites=true&w=majority&appName=vibration-sensor")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	Client = client
	fmt.Println("Connected to MongoDB!")
	return nil
}

func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database("vibration-sensor").Collection(collectionName)
}

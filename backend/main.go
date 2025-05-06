package main

import (
	"log"
	"os"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize MongoDB connection
	err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Initialize default warnings
	err = controllers.InitializeWarnings()
	if err != nil {
		log.Fatal("Failed to initialize warnings:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Sensor Management Routes
	// Handles CRUD operations for vibration sensors
	r.POST("/sensors", controllers.CreateSensor)                        // Create new sensor
	r.POST("/sensors/batch-register", controllers.BatchRegisterSensors) // Batch register sensors
	r.GET("/sensors", controllers.GetSensors)                           // Get all sensors
	r.GET("/sensors/:id", controllers.GetSensor)                        // Get specific sensor
	r.PUT("/sensors/:id", controllers.UpdateSensor)                     // Update sensor
	r.DELETE("/sensors/:id", controllers.DeleteSensor)                  // Delete sensor
	r.POST("/sensors/register", controllers.RegisterSensor)             // Register sensor and get token

	// User Management Routes
	// Handles user registration, authentication, and management
	r.POST("/users", controllers.CreateUser)                        // Register new user
	r.POST("/users/batch-register", controllers.BatchRegisterUsers) // Batch register users
	r.GET("/users", controllers.GetUsers)                           // Get all users
	r.GET("/users/:id", controllers.GetUser)                        // Get specific user
	r.PUT("/users/:id", controllers.UpdateUser)                     // Update user
	r.DELETE("/users/:id", controllers.DeleteUser)                  // Delete user

	r.POST("/login", controllers.Login)                // User login
	r.POST("/refresh-token", controllers.RefreshToken) // Refresh access token

	// Warning Management Routes
	// Handles retrieval of warning information
	r.GET("/warnings", controllers.GetWarnings)    // Get all warnings
	r.GET("/warnings/:id", controllers.GetWarning) // Get specific warning

	// Vibration Data Routes
	r.POST("/vibrations", controllers.CreateVibration)
	r.POST("/vibrations/batch-register", controllers.BatchRegisterVibrations)
	r.GET("/vibrations", controllers.GetVibrations)
	r.GET("/vibrations/:id", controllers.GetVibration)
	r.PUT("/vibrations/:id", controllers.UpdateVibration)
	r.DELETE("/vibrations/:id", controllers.DeleteVibration)

	// Health Check Routes
	// Basic endpoints to check server status
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"Server": "Running"})
	})

	r.GET("/Bruh", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Server Configuration
	// Set port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run("0.0.0.0:" + port)
}

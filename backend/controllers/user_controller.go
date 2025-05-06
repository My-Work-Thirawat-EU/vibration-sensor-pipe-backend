package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/ThirawatEu/vibration-sensor-gas-pipe/config"
	"github.com/ThirawatEu/vibration-sensor-gas-pipe/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func generateTokens(userID primitive.ObjectID) (string, string, time.Time, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.Hex(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours expiration
	})

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.Hex(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days expiration
	})

	// Sign the tokens
	accessTokenString, err := accessToken.SignedString([]byte(config.GetConfig().JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(config.GetConfig().JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessTokenString, refreshTokenString, time.Now().Add(time.Hour * 24), nil
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	user.Password = string(hashedPassword)

	// Generate tokens
	accessToken, refreshToken, tokenExpiry, err := generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
		return
	}

	user.Token = accessToken
	user.RefreshToken = refreshToken
	user.TokenExpiry = tokenExpiry

	collection := config.GetCollection("users")
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	// Don't send password back
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

func GetUsers(c *gin.Context) {
	var users []models.User
	collection := config.GetCollection("users")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user models.User
		cursor.Decode(&user)
		// Don't send passwords back
		user.Password = ""
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var user models.User
	collection := config.GetCollection("users")
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Don't send password back
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the new password if it's being updated
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	collection := config.GetCollection("users")
	update := bson.M{
		"$set": bson.M{
			"username": user.Username,
			"password": user.Password,
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
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := config.GetCollection("users")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Login failed"})
		return
	}

	var user models.User
	collection := config.GetCollection("users")
	err := collection.FindOne(context.Background(), bson.M{"username": loginData.Username}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password", "message": "Login failed"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password", "message": "Login failed"})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, tokenExpiry, err := generateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens", "message": "Login failed"})
		return
	}

	// Update user with new tokens
	update := bson.M{
		"$set": bson.M{
			"token":         accessToken,
			"refresh_token": refreshToken,
			"token_expiry":  tokenExpiry,
		},
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user tokens", "message": "Login failed"})
		return
	}

	// Don't send password back
	user.Password = ""
	user.Token = accessToken
	user.RefreshToken = refreshToken
	user.TokenExpiry = tokenExpiry

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "Login success",
	})
}

func RefreshToken(c *gin.Context) {
	var refreshData struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&refreshData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify refresh token
	token, err := jwt.Parse(refreshData.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWTSecret), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, tokenExpiry, err := generateTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
		return
	}

	// Update user with new tokens
	collection := config.GetCollection("users")
	update := bson.M{
		"$set": bson.M{
			"token":         accessToken,
			"refresh_token": refreshToken,
			"token_expiry":  tokenExpiry,
		},
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		update,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"expires_at":    tokenExpiry,
	})
}

func BatchRegisterUsers(c *gin.Context) {
	var users []models.User
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := config.GetCollection("users")
	var results []models.User
	var errors []string

	for _, user := range users {
		// Hash the password before storing
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			errors = append(errors, "Error hashing password for user: "+user.Username)
			continue
		}
		user.Password = string(hashedPassword)

		// Generate tokens
		accessToken, refreshToken, tokenExpiry, err := generateTokens(user.ID)
		if err != nil {
			errors = append(errors, "Error generating tokens for user: "+user.Username)
			continue
		}

		user.Token = accessToken
		user.RefreshToken = refreshToken
		user.TokenExpiry = tokenExpiry

		result, err := collection.InsertOne(context.Background(), user)
		if err != nil {
			errors = append(errors, "Error creating user: "+user.Username)
			continue
		}

		user.ID = result.InsertedID.(primitive.ObjectID)
		// Don't send password back
		user.Password = ""
		results = append(results, user)
	}

	response := gin.H{
		"successful_registrations": len(results),
		"failed_registrations":     len(errors),
		"users":                    results,
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

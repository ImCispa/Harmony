package handlers

import (
	"context"
	"harmony/db"
	"harmony/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateServer gestisce la creazione di un nuovo server
func CreateServer(c *gin.Context) {
	var server models.Server

	// Validazione input JSON
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Imposta un contesto per MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ottieni la collezione "servers"
	collection := db.Client.Database("harmony").Collection("servers")

	// Inserisce il documento
	result, err := collection.InsertOne(ctx, bson.M{
		"name":        server.Name,
		"description": server.Description,
		"owner_id":    server.OwnerID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          result.InsertedID,
		"name":        server.Name,
		"description": server.Description,
		"owner_id":    server.OwnerID,
	})
}

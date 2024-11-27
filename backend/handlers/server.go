package handlers

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"harmony/db"
	"harmony/models"
	"harmony/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateServer(c *gin.Context) {
	var server models.Server

	// Validate input
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check server codes to get the one to use for the new server
	cServerCodes := db.Client.Database("harmony").Collection("server_codes")
	var serverCode models.ServerCode
	newServerName := false
	err := cServerCodes.FindOne(ctx, bson.M{"name": server.Name}).Decode(&serverCode)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check server name"})
			return
		}
		newServerName = true
	}

	newCode := utils.GetRandomCode(serverCode.Codes)
	if newServerName {
		_, err := cServerCodes.InsertOne(ctx, bson.M{
			"name": 	server.Name,
			"codes": 	[]int{newCode},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server codes"})
			return
		}
	} else {
		update := bson.M{
	        "$set": bson.M{
	            "codes":	append(serverCode.Codes, newCode),
	        },
	    }
		_, err = cServerCodes.UpdateByID(ctx, serverCode.ID, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server codes"})
			return
		}
	}

	// creates new server
	uniqueName := fmt.Sprintf("%s#%d", server.Name, newCode)
	cServers := db.Client.Database("harmony").Collection("servers")
	result, err := cServers.InsertOne(ctx, bson.M{
		"name":        	server.Name,
		"image": 		server.Image,
		"owner_id":    	server.OwnerID,
		"unique_name": 	uniqueName,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":			result.InsertedID,
		"name":        	server.Name,
		"image": 		server.Image,
		"owner_id":		server.OwnerID,
		"unique_name":	uniqueName,
	})
}

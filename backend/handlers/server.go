package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"harmony/db"
	"harmony/models"
	"harmony/utils"
	"net/http"
	"time"
)

var collectionServers = "servers"
var collectionServerCodes = "server_codes"

func CreateServer(c *gin.Context) {
	var server models.Server

	// validate input
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check server codes to get the one to use for the new server
	cServerCodes := db.Client.Database(db.Database).Collection(collectionServerCodes)
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
			"name": server.Name,
			"codes": []int{newCode},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server codes"})
			return
		}
	} else {
		update := bson.M{
	        "$set": bson.M{
	            "codes": append(serverCode.Codes, newCode),
	        },
	    }
		_, err = cServerCodes.UpdateByID(ctx, serverCode.ID, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server codes"})
			return
		}
	}

	// creates new server
	server.GenerateUniqueName(newCode)
	cServers := db.Client.Database(db.Database).Collection(collectionServers)
	result, err := cServers.InsertOne(ctx, bson.M{
		"name": server.Name,
		"image": server.Image,
		"owner_id": server.OwnerID,
		"unique_name": server.UniqueName,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.InsertedID,
		"name": server.Name,
		"image": server.Image,
		"owner_id": server.OwnerID,
		"unique_name": server.UniqueName,
	})
}

func ReadServer(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// search server
	cServer := db.Client.Database(db.Database).Collection(collectionServers)
	var server models.Server
	err = cServer.FindOne(ctx, bson.M{"_id": objectId}).Decode(&server)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": server.ID,
		"name": server.Name,
		"image": server.Image,
		"owner_id": server.OwnerID,
		"unique_name": server.UniqueName,
	})
}

func UpdateServer(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var in models.Server
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// search server
	cServer := db.Client.Database(db.Database).Collection(collectionServers)
	var server models.Server
	err = cServer.FindOne(ctx, bson.M{"_id": objectId}).Decode(&server)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	// update data
	server.Name = in.Name
	server.Image = in.Image
	server.OwnerID = in.OwnerID

	update := bson.M{
        "$set": bson.M{
            "name": server.Name,
			"image": server.Image,
			"owner_id": server.OwnerID,
        },
    }
	_, err = cServer.UpdateByID(ctx, objectId, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": server.ID,
		"name": server.Name,
		"image": server.Image,
		"owner_id": server.OwnerID,
		"unique_name": server.UniqueName,
	})
}

func DeleteServer(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// try deleting
	cServer := db.Client.Database(db.Database).Collection(collectionServers)
	r, err := cServer.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	if r.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	c.Status(http.StatusNoContent)
}
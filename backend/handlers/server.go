package handlers

import (
	"context"
	"errors"
	"fmt"
	"harmony/db"
	"harmony/models"
	"harmony/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	if len(server.Name) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check user exist
	cUser := db.Client.Database(db.Database).Collection(collectionUsers)
	var user models.User
	err := cUser.FindOne(ctx, bson.M{"unique_name": server.OwnerID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// check server codes to get the one to use for the new server
	cServerCodes := db.Client.Database(db.Database).Collection(collectionServerCodes)
	var serverCode models.ServerCode
	newServerName := false
	err = cServerCodes.FindOne(ctx, bson.M{"name": server.Name}).Decode(&serverCode)
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
			"name":  server.Name,
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
		"name":        server.Name,
		"image":       server.Image,
		"owner_id":    server.OwnerID,
		"unique_name": server.UniqueName,
		"users":       []string{server.OwnerID},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"servers": append(user.Servers, server.UniqueName),
		},
	}
	_, err = cUser.UpdateByID(ctx, user.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          result.InsertedID,
		"name":        server.Name,
		"image":       server.Image,
		"owner_id":    server.OwnerID,
		"unique_name": server.UniqueName,
		"users":       []string{server.OwnerID},
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
		"id":          server.ID,
		"name":        server.Name,
		"image":       server.Image,
		"owner_id":    server.OwnerID,
		"unique_name": server.UniqueName,
		"users":       server.Users,
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
			"name":     server.Name,
			"image":    server.Image,
			"owner_id": server.OwnerID,
		},
	}
	_, err = cServer.UpdateByID(ctx, objectId, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          server.ID,
		"name":        server.Name,
		"image":       server.Image,
		"owner_id":    server.OwnerID,
		"unique_name": server.UniqueName,
		"users":       server.Users,
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

func InviteServer(c *gin.Context) {
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

	// compose url
	url := fmt.Sprintf("%s/servers/%s/join?t=%d", utils.GetFullHost(c), server.ID.Hex(), time.Now().Add(5*time.Minute).UnixMilli())

	c.JSON(http.StatusOK, gin.H{
		"link": url,
	})
}

func JoinServer(c *gin.Context) {
	id := c.Param("id")
	tOriginale := c.DefaultQuery("t", "")
	// todo: need to use auth
	userName := c.GetHeader("x-harmony-username")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := strconv.ParseInt(tOriginale, 10, 64)
	if tOriginale == "" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to resolve t"})
		return
	}
	if time.Now().UnixMilli() > t {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Join invite expired"})
		return
	}
	if userName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user"})
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

	// search user
	cUser := db.Client.Database(db.Database).Collection(collectionUsers)
	var user models.User
	err = cUser.FindOne(ctx, bson.M{"unique_name": userName}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// check already in the list
	if utils.Contains(server.Users, user.UniqueName) || utils.Contains(user.Servers, server.UniqueName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already joined server"})
		return
	}

	// todo: need to take user from access token
	server.Users = append(server.Users, user.UniqueName)
	user.Servers = append(user.Servers, server.UniqueName)

	update := bson.M{
		"$set": bson.M{
			"users": server.Users,
		},
	}
	_, err = cServer.UpdateByID(ctx, server.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	update = bson.M{
		"$set": bson.M{
			"servers": user.Servers,
		},
	}
	_, err = cUser.UpdateByID(ctx, user.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func LeaveServer(c *gin.Context) {
	id := c.Param("id")
	// todo: need to use auth
	userName := c.GetHeader("x-harmony-username")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if userName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user"})
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

	// search user
	cUser := db.Client.Database(db.Database).Collection(collectionUsers)
	var user models.User
	err = cUser.FindOne(ctx, bson.M{"unique_name": userName}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	if userName == server.OwnerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner cannot leave server"})
		return
	}

	// check already in the list
	if !utils.Contains(server.Users, user.UniqueName) && !utils.Contains(user.Servers, server.UniqueName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not in the server"})
		return
	}

	// todo: need to take user from access token
	server.Users = utils.Remove(server.Users, user.UniqueName)
	user.Servers = utils.Remove(user.Servers, server.UniqueName)

	update := bson.M{
		"$set": bson.M{
			"users": server.Users,
		},
	}
	_, err = cServer.UpdateByID(ctx, server.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	update = bson.M{
		"$set": bson.M{
			"servers": user.Servers,
		},
	}
	_, err = cUser.UpdateByID(ctx, user.ID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusNoContent)
}

package user

import (
	"context"
	"errors"
	"harmony/internal/database"
	"harmony/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	DB *mongo.Client
}

func NewHandler(db *database.Service) *Handler {
	return &Handler{DB: db.Mongo}
}

var databaseName = "harmony"
var collectionUsers = "users"
var collectionUserCodes = "user_codes"

func (h *Handler) Create(c *gin.Context) {
	var user User

	// validate input
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(user.Name) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name required"})
		return
	}
	if len(user.Mail) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail required"})
		return
	}
	if !utils.IsValidEmail(user.Mail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail not valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check mail to see if already used
	cUsers := h.DB.Database(databaseName).Collection(collectionUsers)
	var uMailUsed User
	err := cUsers.FindOne(ctx, bson.M{"mail": user.Mail}).Decode(&uMailUsed)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail already used"})
		return
	} else if uMailUsed.Mail == user.Mail {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail already used"})
		return
	}

	// check user codes to get the one to use for the new user
	cUserCodes := h.DB.Database(databaseName).Collection(collectionUserCodes)
	var userCode UserCode
	newUserName := false
	err = cUserCodes.FindOne(ctx, bson.M{"name": user.Name}).Decode(&userCode)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user name"})
			return
		}
		newUserName = true
	}

	newCode := utils.GetRandomCode(userCode.Codes)
	if newUserName {
		_, err := cUserCodes.InsertOne(ctx, bson.M{
			"name":  user.Name,
			"codes": []int{newCode},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user codes"})
			return
		}
	} else {
		update := bson.M{
			"$set": bson.M{
				"codes": append(userCode.Codes, newCode),
			},
		}
		_, err = cUserCodes.UpdateByID(ctx, userCode.ID, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user codes"})
			return
		}
	}

	// creates new user
	user.GenerateUniqueName(newCode)
	result, err := cUsers.InsertOne(ctx, bson.M{
		"name":        user.Name,
		"mail":        user.Mail,
		"unique_name": user.UniqueName,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          result.InsertedID,
		"name":        user.Name,
		"mail":        user.Mail,
		"unique_name": user.UniqueName,
		"servers":     user.Servers,
	})
}

func (h *Handler) Read(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// search user
	cUser := h.DB.Database(databaseName).Collection(collectionUsers)
	var user User
	err = cUser.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"mail":        user.Mail,
		"unique_name": user.UniqueName,
		"servers":     user.Servers,
	})
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var in User
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// search user
	cUser := h.DB.Database(databaseName).Collection(collectionUsers)
	var user User
	err = cUser.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// update data
	user.Name = in.Name

	update := bson.M{
		"$set": bson.M{
			"name": user.Name,
		},
	}
	_, err = cUser.UpdateByID(ctx, objectId, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"mail":        user.Mail,
		"unique_name": user.UniqueName,
		"servers":     user.Servers,
	})
}

func (h *Handler) Delete(c *gin.Context) {
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
	cUser := h.DB.Database(databaseName).Collection(collectionUsers)
	r, err := cUser.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if r.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	c.Status(http.StatusNoContent)
}

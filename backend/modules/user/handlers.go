package user

import (
	"context"
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
	DB   *mongo.Client
	Repo *Repository
}

func NewHandler(db *database.Service) *Handler {
	return &Handler{
		DB:   db.Mongo,
		Repo: NewRepository(db.Mongo),
	}
}

var databaseName = "harmony"
var collectionUsers = "users"

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

	// check mail to see if already used
	isMailUsed, err := h.Repo.IsMailUsed(user.Mail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Falied"})
		return
	}
	if isMailUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail already used"})
		return
	}

	// create
	err = h.Repo.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          user.ID,
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

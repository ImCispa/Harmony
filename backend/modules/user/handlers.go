package user

import (
	"harmony/internal/database"
	"harmony/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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
	err = h.Repo.Create(&user)
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

	// search user
	user, err := h.Repo.Read(objectId)
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

	// search user
	user, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// update data
	user.Name = in.Name

	err = h.Repo.Update(*user)
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

	// try deleting
	isDeleted, err := h.Repo.Delete(objectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if !isDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	c.Status(http.StatusNoContent)
}

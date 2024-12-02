package user

import (
	"harmony/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(db *database.Service) *Handler {
	return &Handler{
		Repo: NewRepository(db.Mongo),
	}
}

func (h *Handler) Create(c *gin.Context) {
	type RequestBody struct {
		Name string `json:"name"`
		Mail string `json:"mail"`
	}
	var rb RequestBody

	// validate input
	if err := c.ShouldBindJSON(&rb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if check, errMsg := IsNameValid(rb.Name); !check {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	if check, errMsg := IsMailValid(rb.Mail); !check {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// check mail to see if already used
	isMailUsed, err := h.Repo.IsMailUsed(rb.Mail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Falied"})
		return
	}
	if isMailUsed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mail already used"})
		return
	}

	// create
	user := NewUser(rb.Name, rb.Mail)

	err = h.Repo.Create(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.Print())
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

	c.JSON(http.StatusOK, user.Print())
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	type RequestBody struct {
		Name string `json:"name"`
	}
	var rb RequestBody

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&rb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if check, errMsg := IsNameValid(rb.Name); !check {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// search user
	user, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// update data
	user.Name = rb.Name

	err = h.Repo.Update(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.Print())
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

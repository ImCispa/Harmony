package server

import (
	"fmt"
	"harmony/internal/database"
	"harmony/modules/user"
	"harmony/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	Repo     *Repository
	RepoUser *user.Repository
}

func NewHandler(db *database.Service) *Handler {
	return &Handler{
		Repo:     NewRepository(db.Mongo),
		RepoUser: user.NewRepository(db.Mongo),
	}
}

func (h *Handler) Create(c *gin.Context) {
	type RequestBody struct {
		Name    string `json:"name"`
		Image   string `json:"image"`
		OwnerID string `json:"owner_id"` // TODO: owner should be retreived from token
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

	// check if user exist
	user, err := h.RepoUser.ReadByUniqueName(rb.OwnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// create
	server := NewServer(rb.Name, rb.Image, rb.OwnerID)

	err = h.Repo.Create(&server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	// insert in user the new server created
	user.Servers = append(user.Servers, server.UniqueName)
	err = h.RepoUser.Update(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusCreated, server.Print())
}

func (h *Handler) Read(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// search server
	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	c.JSON(http.StatusOK, server.Print())
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	type RequestBody struct {
		Name  string `json:"name"`
		Image string `json:"image"`
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

	// search server
	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	// update data
	server.Name = rb.Name
	server.Image = rb.Image

	err = h.Repo.Update(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	c.JSON(http.StatusOK, server.Print())
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

func (h *Handler) Invite(c *gin.Context) {
	id := c.Param("id")

	// validate input
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// search server
	server, err := h.Repo.Read(objectId)
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

func (h *Handler) Join(c *gin.Context) {
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

	// search server
	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	// search user
	user, err := h.RepoUser.ReadByUniqueName(userName)
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

	err = h.Repo.Update(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	err = h.RepoUser.Update(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Leave(c *gin.Context) {
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

	// search server
	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	// search user
	user, err := h.RepoUser.ReadByUniqueName(userName)
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

	err = h.Repo.Update(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	err = h.RepoUser.Update(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusNoContent)
}

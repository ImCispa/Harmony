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
		Name  string `json:"name"`
		Image string `json:"image"`
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
	sub, ok := utils.GetSub(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retreive user"})
		return
	}
	user, err := h.RepoUser.ReadByUniqueName(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// create
	server := NewServer(rb.Name, rb.Image, user.UniqueName)

	err = h.Repo.Create(&server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	// insert in user the new server created
	if user.Servers == nil {
		user.Servers = map[string]string{
			server.UniqueName: "owner",
		}
	} else {
		user.Servers[server.UniqueName] = "owner"
	}
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

	// role check
	if utils.IsRoleAtLeastAdmin(c, server.UniqueName) {
		c.Status(http.StatusUnauthorized)
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

	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive server"})
		return
	}

	// role check
	if utils.IsRoleAdmin(c, server.UniqueName) {
		c.Status(http.StatusUnauthorized)
		return
	}

	// try deleting
	isDeleted, err := h.Repo.Delete(objectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}
	if !isDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
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

	// role check
	if utils.IsRoleAtLeastMember(c, server.UniqueName) {
		c.Status(http.StatusUnauthorized)
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

	sub, ok := utils.GetSub(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retreive user"})
		return
	}

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

	// search server
	server, err := h.Repo.Read(objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive server"})
		return
	}

	// search user
	user, err := h.RepoUser.ReadByUniqueName(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	// check already in the list
	_, existsInServer := server.Users[user.UniqueName]

	if existsInServer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already joined server"})
		return
	}

	// todo: need to take user from access token
	server.Users[user.UniqueName] = "member"
	if user.Servers == nil {
		user.Servers = map[string]string{
			server.UniqueName: "member",
		}
	} else {
		user.Servers[server.UniqueName] = "owner"
	}

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

	sub, ok := utils.GetSub(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retreive user"})
		return
	}

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

	// search user
	user, err := h.RepoUser.ReadByUniqueName(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to retreive user"})
		return
	}

	if user.UniqueName == server.OwnerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner cannot leave server"})
		return
	}

	// check already in the list
	_, existsInServer := server.Users[user.UniqueName]

	if !existsInServer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not in the server"})
		return
	}

	// todo: need to take user from access token
	delete(server.Users, user.UniqueName)
	delete(user.Servers, server.UniqueName)

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

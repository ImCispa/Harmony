package server

import (
	"net/http"

	"harmony/modules/server"
	"harmony/modules/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/health", s.healthHandler)

	serverHandler := server.NewHandler(&s.db)
	serverGroup := r.Group("/servers")
	server.RegisterRoutes(serverGroup, serverHandler)

	userHandler := user.NewHandler(&s.db)
	userGroup := r.Group("/users")
	user.RegisterRoutes(userGroup, userHandler)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

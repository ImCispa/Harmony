package server

import (
	"context"
	"net/http"
	"strings"

	"harmony/modules/server"
	"harmony/modules/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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
	r.GET("/login", s.loginHandler)
	r.GET("/callback", s.callbackHandler)

	serverHandler := server.NewHandler(&s.db)
	serverGroup := r.Group("/servers", s.authMiddleware())
	server.RegisterRoutes(serverGroup, serverHandler)

	userHandler := user.NewHandler(&s.db)
	userGroup := r.Group("/users", s.authMiddleware())
	user.RegisterRoutes(userGroup, userHandler)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) loginHandler(c *gin.Context) {
	// Genera l'URL per il login
	url := s.auth.Oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func (s *Server) callbackHandler(c *gin.Context) {
	ctx := context.Background()

	// Verifica lo stato (se necessario)
	state := c.Query("state")
	if state != "state" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	token, err := s.auth.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token", "details": err.Error()})
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token in token response"})
		return
	}

	//idToken, err := s.auth.OidcVerifierOauth.Verify(ctx, rawIDToken)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID Token", "details": err.Error()})
	//	return
	//}

	//var claims map[string]interface{}
	//if err := idToken.Claims(&claims); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse claims", "details": err.Error()})
	//	return
	//}

	//email := claims["email"].(string)

	c.SetCookie("token", rawIDToken, 3600, "/", s.host, false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": token.AccessToken,
		"id_token":     rawIDToken,
	})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		idToken, err := s.auth.OidcVerifier.Verify(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		var claims map[string]interface{}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse token claims"})
			return
		}

		// Passa il controllo al prossimo handler
		c.Set("claims", claims)
		c.Next()
	}
}

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"harmony/modules/server"
	"harmony/modules/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	userRegGroup := r.Group("/user/registration")
	user.RegisterRoutesNoAuth(userRegGroup, userHandler)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token in token response"})
		return
	}

	idToken, err := s.auth.OidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID Token"})
		return
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse claims"})
		return
	}

	email := claims["email"].(string)

	userRepo := user.NewRepository(s.db.Mongo)
	user, err := userRepo.ReadByMail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed in get user"})
		return
	}

	newClaims := jwt.MapClaims{
		"sub":   user.UniqueName,
		"exp":   time.Now().Add(time.Minute * 15).Unix(),
		"aud":   s.auth.Oauth2Config.ClientID,
		"nbf":   time.Now().Add(time.Minute * -5).Unix(),
		"iat":   time.Now().Unix(),
		"roles": user.Servers,
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)

	signedToken, err := newToken.SignedString([]byte(s.auth.Oauth2Config.ClientSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed in token sign"})
		return
	}

	c.SetCookie("token", signedToken, 3600, "/", s.host, false, true)

	c.JSON(http.StatusOK, gin.H{
		"id_token": signedToken,
	})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenString := authHeader[len("Bearer "):]

		// Parse e verifica il token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Controlla il metodo di firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Method)
			}
			// Restituisci la chiave usata per verificare
			return []byte(s.auth.Oauth2Config.ClientSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Recupera i claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		// Passa il controllo al prossimo handler
		c.Set("claims", claims)
		c.Next()
	}
}

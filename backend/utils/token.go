package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetSub(c *gin.Context) (string, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return "", false
	}
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	sub, ok := mapClaims["sub"].(string)
	if !ok {
		return "", false
	}
	return sub, true
}

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

func GetRole(c *gin.Context, serverUniqueName string) (string, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return "", false
	}
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	roles, ok := mapClaims["roles"].(map[string]string)
	if !ok {
		return "", false
	}
	if roles == nil {
		return "", false
	}
	role, exists := roles[serverUniqueName]
	if !exists {
		return "", false
	}
	return role, true
}

func IsRoleOwner(c *gin.Context, serverUniqueName string) bool {
	role, ok := GetRole(c, serverUniqueName)
	if !ok {
		return false
	}
	if role != "owner" {
		return false
	}
	return true
}

func IsRoleAdmin(c *gin.Context, serverUniqueName string) bool {
	role, ok := GetRole(c, serverUniqueName)
	if !ok {
		return false
	}
	if role != "admin" {
		return false
	}
	return true
}

func IsRoleAtLeastAdmin(c *gin.Context, serverUniqueName string) bool {
	role, ok := GetRole(c, serverUniqueName)
	if !ok {
		return false
	}
	if role != "owner" && role != "admin" {
		return false
	}
	return true
}

func IsRoleMember(c *gin.Context, serverUniqueName string) bool {
	role, ok := GetRole(c, serverUniqueName)
	if !ok {
		return false
	}
	if role != "member" {
		return false
	}
	return true
}

func IsRoleAtLeastMember(c *gin.Context, serverUniqueName string) bool {
	role, ok := GetRole(c, serverUniqueName)
	if !ok {
		return false
	}
	if role != "owner" && role != "admin" && role != "member" {
		return false
	}
	return true
}

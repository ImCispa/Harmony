package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetFullHost(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	fullHost := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return fullHost
}

func GetFullUrl(c *gin.Context) string {
	return  fmt.Sprintf("%s%s", GetFullHost(c), c.Request.URL.String())
}

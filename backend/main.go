package main

import (
	"github.com/gin-gonic/gin"
	"harmony/db"
	"harmony/handlers"
)

func main() {
	// checks db connection
	db.Init()

	router := gin.Default()

	router.POST("/servers", handlers.CreateServer)

	router.Run(":8080")
}

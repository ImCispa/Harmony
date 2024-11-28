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
	router.GET("/servers/:id", handlers.ReadServer)
	router.PATCH("/servers/:id", handlers.UpdateServer)
	router.DELETE("/servers/:id", handlers.DeleteServer)

	router.POST("/users", handlers.CreateUser)
	router.GET("/users/:id", handlers.ReadUser)
	router.PATCH("/users/:id", handlers.UpdateUser)
	router.DELETE("/users/:id", handlers.DeleteUser)

	router.Run(":8080")
}

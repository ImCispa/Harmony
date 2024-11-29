package main

import (
	"harmony/db"
	"harmony/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// checks db connection
	db.Init()

	router := gin.Default()

	router.POST("/servers", handlers.CreateServer)
	router.GET("/servers/:id", handlers.ReadServer)
	router.PATCH("/servers/:id", handlers.UpdateServer)
	router.DELETE("/servers/:id", handlers.DeleteServer)
	router.GET("/servers/:id/invite", handlers.InviteServer)
	router.POST("/servers/:id/join", handlers.JoinServer)
	router.POST("/servers/:id/leave", handlers.LeaveServer)

	router.POST("/users", handlers.CreateUser)
	router.GET("/users/:id", handlers.ReadUser)
	router.PATCH("/users/:id", handlers.UpdateUser)
	router.DELETE("/users/:id", handlers.DeleteUser)

	router.Run(":8080")
}

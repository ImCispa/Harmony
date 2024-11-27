package main

import (
	"github.com/gin-gonic/gin"
	"harmony/db"
	"harmony/handlers"
)

func main() {
	// Inizializza il database
	db.Init()

	// Crea un'istanza del router
	router := gin.Default()

	// Definisci le rotte per i server
	router.POST("/servers", handlers.CreateServer)

	// Avvia il server
	router.Run(":8080") // Il server sarà disponibile su localhost:8080
}

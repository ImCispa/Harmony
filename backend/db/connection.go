package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
)

var Client *mongo.Client

func Init() {
	// Carica le variabili d'ambiente
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Recupera la stringa di connessione
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatalf("MONGO_URI not set in .env file")
	}

	// Configura il client di MongoDB
	clientOptions := options.Client().ApplyURI(uri)

	// Connessione al database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Verifica la connessione
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	fmt.Println("MongoDB connection established")
}

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Server rappresenta la struttura di un server nel database MongoDB
type Server struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // Generato automaticamente
	Name        string             `bson:"name" json:"name" binding:"required"`
	Description string             `bson:"description" json:"description"`
	OwnerID     string             `bson:"owner_id" json:"owner_id"`
}

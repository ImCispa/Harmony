package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name" json:"name" binding:"required"`
	UniqueCode string `bson:"unique_code" json:"unique_code" binding:"required"`
	UniqueName string `bson:"unique_name"`
}
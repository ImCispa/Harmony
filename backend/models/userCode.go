package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserCode struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name"`
	Codes []int `bson:"codes"`
}

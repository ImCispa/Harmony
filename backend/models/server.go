package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Server struct {
	ID          primitive.ObjectID	`bson:"_id,omitempty"`
	Name        string             	`bson:"name" json:"name" binding:"required"`
	Image 		string 			   	`bson:"image" json:"image"`
	OwnerID     string             	`bson:"owner_id" json:"owner_id"`
	UniqueName	string				`bson:"unique_name"`
}

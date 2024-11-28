package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Server struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name" json:"name"`
	Image string `bson:"image" json:"image"`
	OwnerID string `bson:"owner_id" json:"owner_id"`
	UniqueName string `bson:"unique_name"`
	Users []string `bson:"users"`
}

func (s *Server) GenerateUniqueName(code int) string {
	s.UniqueName = fmt.Sprintf("%s:%04d", s.Name, code)
	return s.UniqueName
}
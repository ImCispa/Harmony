package server

import (
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Server struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Image      string             `bson:"image"`
	OwnerID    string             `bson:"owner_id"`
	UniqueName string             `bson:"unique_name"`
	Users      map[string]string  `bson:"users"`
}

func NewServer(name string, image string, ownerId string) Server {
	return Server{
		Name:    name,
		Image:   image,
		OwnerID: ownerId,
		Users: map[string]string{
			ownerId: "owner",
		},
	}
}

func IsNameValid(name string) (bool, string) {
	if len(name) == 0 {
		return false, "Name is empty"
	}
	var pattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !pattern.MatchString(name) {
		return false, "Name contains not allowed caracters"
	}
	return true, ""
}

func (s *Server) GenerateUniqueName(code int) {
	if s.ID != primitive.NilObjectID {
		return
	}
	s.UniqueName = fmt.Sprintf("%s:%04d", s.Name, code)
}

func (s *Server) Print() map[string]any {
	return map[string]any{
		"id":          s.ID,
		"name":        s.Name,
		"image":       s.Image,
		"unique_name": s.UniqueName,
		"users":       s.Users,
	}
}

type ServerCode struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Codes []int              `bson:"codes"`
}

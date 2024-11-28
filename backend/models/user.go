package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name" json:"name" binding:"required"`
	Mail string `bson:"mail" json:"mail" binding:"required"`
	UniqueName string `bson:"unique_name"`
}

func (u *User) GenerateUniqueName(code int) string {
	u.UniqueName = fmt.Sprintf("%s:%04d", u.Name, code)
	return u.UniqueName
}
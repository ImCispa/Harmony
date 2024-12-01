package user

import (
	"fmt"
	"harmony/utils"
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name" json:"name"`
	Mail       string             `bson:"mail" json:"mail"`
	UniqueName string             `bson:"unique_name"`
	Servers    []string           `bson:"servers"`
}

func NewUser(name string, mail string) User {
	return User{
		Name: name,
		Mail: mail,
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

func IsMailValid(mail string) (bool, string) {
	if len(mail) == 0 {
		return false, "Mail is empty"
	}
	if !utils.IsValidEmail(mail) {
		return false, "Mail is not in the right format"
	}
	return true, ""
}

func (u *User) GenerateUniqueName(code int) {
	if u.ID != primitive.NilObjectID {
		return
	}
	u.UniqueName = fmt.Sprintf("%s:%04d", u.Name, code)
}

func (u *User) Print() map[string]any {
	return map[string]any{
		"id":          u.ID,
		"name":        u.Name,
		"mail":        u.Mail,
		"unique_name": u.UniqueName,
		"servers":     u.Servers,
	}
}

type UserCode struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Codes []int              `bson:"codes"`
}

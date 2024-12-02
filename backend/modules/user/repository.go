package user

import (
	"context"
	"errors"
	"harmony/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultTimeout = 5 * time.Second

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Client) *Repository {
	return &Repository{db: db.Database("harmony")}
}

func (r *Repository) IsMailUsed(mail string) (bool, error) {
	cUsers := r.db.Collection("users")
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	filter := bson.M{
		"mail": mail,
	}
	err := cUsers.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) Create(user *User) error {
	cUserCodes := r.db.Collection("user_codes")
	cUsers := r.db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var userCode UserCode
	newUserName := false
	filter := bson.M{
		"name": user.Name,
	}
	err := cUserCodes.FindOne(ctx, filter).Decode(&userCode)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		newUserName = true
	}

	newCode := utils.GetRandomCode(userCode.Codes)
	if newUserName {
		_, err := cUserCodes.InsertOne(ctx, bson.M{
			"name":  user.Name,
			"codes": []int{newCode},
		})
		if err != nil {
			return err
		}
	} else {
		update := bson.M{
			"$set": bson.M{
				"codes": append(userCode.Codes, newCode),
			},
		}
		_, err = cUserCodes.UpdateByID(ctx, userCode.ID, update)
		if err != nil {
			return err
		}
	}

	// creates new user
	user.GenerateUniqueName(newCode)
	result, err := cUsers.InsertOne(ctx, bson.M{
		"name":        user.Name,
		"mail":        user.Mail,
		"unique_name": user.UniqueName,
	})
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (r *Repository) Read(id primitive.ObjectID) (*User, error) {
	cUsers := r.db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User
	err := cUsers.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) ReadByUniqueName(uniqueName string) (*User, error) {
	cUsers := r.db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User
	err := cUsers.FindOne(ctx, bson.M{"unique_name": uniqueName}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) Update(user *User) error {
	cUsers := r.db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// cannot update:
	// - mail
	// - unique name
	update := bson.M{
		"$set": bson.M{
			"name":    user.Name,
			"servers": user.Servers,
		},
	}
	_, err := cUsers.UpdateByID(ctx, user.ID, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(id primitive.ObjectID) (bool, error) {
	cUsers := r.db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := cUsers.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}

	if result.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

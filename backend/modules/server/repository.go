package server

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
	return &Repository{
		db: db.Database("harmony"),
	}
}

func (r *Repository) Create(server *Server) error {
	cServerCodes := r.db.Collection("server_codes")
	cServers := r.db.Collection("servers")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var serverCode ServerCode
	newServerName := false
	filter := bson.M{
		"name": server.Name,
	}
	err := cServerCodes.FindOne(ctx, filter).Decode(&serverCode)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		newServerName = true
	}

	newCode := utils.GetRandomCode(serverCode.Codes)
	if newServerName {
		_, err := cServerCodes.InsertOne(ctx, bson.M{
			"name":  server.Name,
			"codes": []int{newCode},
		})
		if err != nil {
			return err
		}
	} else {
		update := bson.M{
			"$set": bson.M{
				"codes": append(serverCode.Codes, newCode),
			},
		}
		_, err = cServerCodes.UpdateByID(ctx, serverCode.ID, update)
		if err != nil {
			return err
		}
	}

	// creates new server
	server.GenerateUniqueName(newCode)
	server.Users = map[string]string{server.OwnerID: "owner"}
	result, err := cServers.InsertOne(ctx, bson.M{
		"name":        server.Name,
		"image":       server.Image,
		"owner_id":    server.OwnerID,
		"unique_name": server.UniqueName,
		"users":       server.Users,
	})
	if err != nil {
		return err
	}

	server.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (r *Repository) Read(id primitive.ObjectID) (*Server, error) {
	cServers := r.db.Collection("servers")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var server Server
	err := cServers.FindOne(ctx, bson.M{"_id": id}).Decode(&server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (r *Repository) Update(server *Server) error {
	cServers := r.db.Collection("servers")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// cannot update:
	// - unique name
	update := bson.M{
		"$set": bson.M{
			"name":     server.Name,
			"image":    server.Image,
			"owner_id": server.OwnerID,
			"users":    server.Users,
		},
	}
	_, err := cServers.UpdateByID(ctx, server.ID, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(id primitive.ObjectID) (bool, error) {
	cServers := r.db.Collection("servers")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := cServers.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}

	if result.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

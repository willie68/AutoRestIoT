package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID          primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Password    string             `json:"password" bson:"password,omitempty"`
	NewPassword string             `json:"newpassword" bson:"-"`
	Admin       bool               `json:"admin" bson:"admin,omitempty"`
	Guest       bool               `json:"guest" bson:"guest,omitempty"`
	Roles       []string           `json:"roles" bson:"roles,omitempty"`
}

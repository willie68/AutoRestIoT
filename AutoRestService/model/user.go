package model

import "go.mongodb.org/mongo-driver/bson/primitive"

//User the user model
type User struct {
	ID          primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Firstname   string             `json:"firstname" bson:"firstname,omitempty"`
	Lastname    string             `json:"lastname" bson:"lastname,omitempty"`
	Salt        []byte             `json:"salt,omitempty" bson:"salt,omitempty"`
	Password    string             `json:"password,omitempty" bson:"password,omitempty"`
	NewPassword string             `json:"newpassword,omitempty" bson:"-"`
	Admin       bool               `json:"admin" bson:"admin,omitempty"`
	Guest       bool               `json:"guest" bson:"guest,omitempty"`
	Roles       []string           `json:"roles" bson:"roles,omitempty"`
}

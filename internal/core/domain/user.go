package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Username      string             `bson:"username,omitempty" validate:"omitempty,min=3,max=30"`
	DisplayName   string             `bson:"displayName,omitempty"`
	Address       string             `bson:"address,omitempty"`
	Email         string             `bson:"email,omitempty" validate:"omitempty,email"`
	Password      string             `bson:"password,omitempty"`
	Citizenship   string             `bson:"citizenship,omitempty"`
	City          string             `bson:"city,omitempty"`
	Nonce         int                `bson:"nonce"`
	SubsidizedOps []string           `bson:"subsidizedOperations"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Address       string             `bson:"address"`
	Email         string             `bson:"email,omitempty"`
	Nonce         int                `bson:"nonce"`
	SubsidizedOps []string           `bson:"subsidizedOperations"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

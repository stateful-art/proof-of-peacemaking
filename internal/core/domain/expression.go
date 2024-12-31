package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Expression struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Creator   primitive.ObjectID `bson:"creator"`
	Content   map[string]string  `bson:"content"`
	IPFSHash  string             `bson:"ipfsHash"`
	OnChainID int                `bson:"onChainId"`
	Status    string             `bson:"status"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

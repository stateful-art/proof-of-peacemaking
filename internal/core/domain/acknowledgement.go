package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Acknowledgement struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ExpressionID primitive.ObjectID `bson:"expressionId"`
	Acknowledger primitive.ObjectID `bson:"acknowledger"`
	Content      map[string]string  `bson:"content"`
	IPFSHash     string             `bson:"ipfsHash"`
	OnChainID    int                `bson:"onChainId"`
	Status       string             `bson:"status"`
	CreatedAt    time.Time          `bson:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"`
}

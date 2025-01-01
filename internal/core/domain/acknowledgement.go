package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AcknowledgementStatus string

const (
	AcknowledgementStatusActive  AcknowledgementStatus = "ACTIVE"
	AcknowledgementStatusRefuted AcknowledgementStatus = "REFUTED"
)

type Acknowledgement struct {
	ID           primitive.ObjectID    `bson:"_id,omitempty"`
	ExpressionID string                `bson:"expressionId"`
	Acknowledger string                `bson:"acknowledger"`
	Content      map[string]string     `bson:"content"`
	IPFSHash     string                `bson:"ipfsHash"`
	OnChainID    int                   `bson:"onChainId"`
	Status       AcknowledgementStatus `bson:"status"`
	CreatedAt    time.Time             `bson:"createdAt"`
	UpdatedAt    time.Time             `bson:"updatedAt"`
}

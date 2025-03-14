package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProofNFT struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	TokenID      int                `bson:"tokenId"`
	Expression   string             `bson:"expressionId"`
	Acknowledger string             `bson:"acknowledgerId"`
	IPFSHash     string             `bson:"ipfsHash"`
	Status       string             `bson:"status"`
	CreatedAt    time.Time          `bson:"createdAt"`
	MintedAt     *time.Time         `bson:"mintedAt,omitempty"`
}

type ProofRequestStatus string

const (
	ProofRequestPending  ProofRequestStatus = "PENDING"
	ProofRequestAccepted ProofRequestStatus = "ACCEPTED"
	ProofRequestRejected ProofRequestStatus = "REJECTED"
)

type ProofRequest struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ExpressionID string             `bson:"expressionId"`
	InitiatorID  string             `bson:"initiatorId"`
	PeerID       string             `bson:"peerId"`
	Status       ProofRequestStatus `bson:"status"`
	CreatedAt    time.Time          `bson:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"`
}

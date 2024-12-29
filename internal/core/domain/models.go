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

type ProofRequest struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ExpressionID string             `bson:"expressionId"`
	InitiatorID  string             `bson:"initiatorId"`
	PeerID       string             `bson:"peerId"`
	Status       ProofRequestStatus `bson:"status"`
	CreatedAt    time.Time          `bson:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"`
}

type ProofRequestStatus string

const (
	ProofRequestPending  ProofRequestStatus = "PENDING"
	ProofRequestAccepted ProofRequestStatus = "ACCEPTED"
	ProofRequestRejected ProofRequestStatus = "REJECTED"
)

type ProofNFT struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	TokenID      int                `bson:"tokenId"`
	Expression   primitive.ObjectID `bson:"expressionId"`
	Acknowledger primitive.ObjectID `bson:"acknowledgerId"`
	IPFSHash     string             `bson:"ipfsHash"`
	Status       string             `bson:"status"`
	CreatedAt    time.Time          `bson:"createdAt"`
	MintedAt     *time.Time         `bson:"mintedAt,omitempty"`
}

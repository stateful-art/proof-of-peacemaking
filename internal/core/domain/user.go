package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Address       string             `bson:"address"`
	Nonce         int                `bson:"nonce"`
	SubsidizedOps []string           `bson:"subsidizedOperations"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

type NotificationType string

const (
	NotificationNewAcknowledgement       NotificationType = "NEW_ACKNOWLEDGEMENT"
	NotificationProofRequestReceived     NotificationType = "PROOF_REQUEST_RECEIVED"
	NotificationProofRequestApproved     NotificationType = "PROOF_REQUEST_APPROVED"
	NotificationNFTMinted                NotificationType = "NFT_MINTED"
	NotificationExpressionConfirmed      NotificationType = "EXPRESSION_CONFIRMED"
	NotificationAcknowledgementConfirmed NotificationType = "ACKNOWLEDGEMENT_CONFIRMED"
)

type Notification struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty"`
	Type      NotificationType       `bson:"type"`
	Title     string                 `bson:"title"`
	Message   string                 `bson:"message"`
	Data      map[string]interface{} `bson:"data"`
	CreatedAt time.Time              `bson:"createdAt"`
}

type UserNotification struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         primitive.ObjectID `bson:"userId"`
	NotificationID primitive.ObjectID `bson:"notificationId"`
	Read           bool               `bson:"read"`
	ReadAt         *time.Time         `bson:"readAt,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt"`
}

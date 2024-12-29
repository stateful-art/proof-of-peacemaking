package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotificationNewAcknowledgement       NotificationType = "NEW_ACKNOWLEDGEMENT"
	NotificationProofRequestReceived     NotificationType = "PROOF_REQUEST_RECEIVED"
	NotificationProofRequestApproved     NotificationType = "PROOF_REQUEST_APPROVED"
	NotificationNFTMinted                NotificationType = "NFT_MINTED"
	NotificationExpressionConfirmed      NotificationType = "EXPRESSION_CONFIRMED"
	NotificationAcknowledgementConfirmed NotificationType = "ACKNOWLEDGEMENT_CONFIRMED"
	NotificationProofRequestAccepted     NotificationType = "PROOF_REQUEST_ACCEPTED"
	NotificationProofRequestRejected     NotificationType = "PROOF_REQUEST_REJECTED"
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

package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	// Existing notification types
	NotificationNewAcknowledgement       NotificationType = "NEW_ACKNOWLEDGEMENT"
	NotificationProofRequestReceived     NotificationType = "PROOF_REQUEST_RECEIVED"
	NotificationProofRequestApproved     NotificationType = "PROOF_REQUEST_APPROVED"
	NotificationNFTMinted                NotificationType = "NFT_MINTED"
	NotificationExpressionConfirmed      NotificationType = "EXPRESSION_CONFIRMED"
	NotificationAcknowledgementConfirmed NotificationType = "ACKNOWLEDGEMENT_CONFIRMED"
	NotificationProofRequestAccepted     NotificationType = "PROOF_REQUEST_ACCEPTED"
	NotificationProofRequestRejected     NotificationType = "PROOF_REQUEST_REJECTED"

	// New conversation-related notification types
	NotificationConversationCreated    NotificationType = "CONVERSATION_CREATED"
	NotificationConversationStarted    NotificationType = "CONVERSATION_STARTED"
	NotificationConversationEnded      NotificationType = "CONVERSATION_ENDED"
	NotificationConversationJoined     NotificationType = "CONVERSATION_JOINED"
	NotificationConversationInvitation NotificationType = "CONVERSATION_INVITATION"
)

type Notification struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID    string                 `bson:"userId" json:"userId"`
	Type      NotificationType       `bson:"type" json:"type"`
	Title     string                 `bson:"title" json:"title"`
	Message   string                 `bson:"message" json:"message"`
	Data      map[string]interface{} `bson:"data" json:"data"`
	Read      bool                   `bson:"read" json:"read"`
	ReadAt    *time.Time             `bson:"readAt,omitempty" json:"readAt,omitempty"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
}

type UserNotification struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         primitive.ObjectID `bson:"userId"`
	NotificationID primitive.ObjectID `bson:"notificationId"`
	Read           bool               `bson:"read"`
	ReadAt         *time.Time         `bson:"readAt,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt"`
}

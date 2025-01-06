package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationStatus string

const (
	ConversationStatusScheduled ConversationStatus = "scheduled"
	ConversationStatusLive      ConversationStatus = "live"
	ConversationStatusEnded     ConversationStatus = "ended"
)

type Conversation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	ImageURL    string             `bson:"imageUrl" json:"imageUrl"`
	CreatorID   string             `bson:"creatorId" json:"creatorId"`
	Status      ConversationStatus `bson:"status" json:"status"`
	StartTime   time.Time          `bson:"startTime" json:"startTime"`
	EndTime     *time.Time         `bson:"endTime,omitempty" json:"endTime,omitempty"`
	Tags        []string           `bson:"tags" json:"tags"`
	RoomName    string             `bson:"roomName" json:"roomName"`

	// LiveKit specific fields
	RoomID    string    `bson:"roomId" json:"roomId"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`

	// Notification subscriptions
	Subscribers []string `bson:"subscribers" json:"subscribers"`
}

package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL       string             `bson:"url" json:"url"`
	AddedBy   string             `bson:"added_by" json:"addedBy"`
	AddedAt   time.Time          `bson:"added_at" json:"addedAt"`
	PlayedAt  *time.Time         `bson:"played_at,omitempty" json:"playedAt,omitempty"`
	IsPlaying bool               `bson:"is_playing" json:"isPlaying"`
}

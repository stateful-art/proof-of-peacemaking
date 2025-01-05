package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Token          string             `bson:"token"`
	UserID         string             `bson:"userId"`
	WebAuthnData   string             `bson:"webauthnData"` // For storing WebAuthn session data
	Address        string             `bson:"address"`
	ExpiresAt      time.Time          `bson:"expiresAt"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
	IsRegistration bool               `bson:"isRegistration"` // Indicates if this is a registration session
}

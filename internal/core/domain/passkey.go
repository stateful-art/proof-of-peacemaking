package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PasskeyCredential represents a WebAuthn credential stored in the database
type PasskeyCredential struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CredentialID []byte             `bson:"credentialId" json:"credentialId"`
	PublicKey    []byte             `bson:"publicKey" json:"publicKey"`
	AAGUID       []byte             `bson:"aaguid" json:"aaguid"`
	SignCount    uint32             `bson:"signCount" json:"signCount"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// UserPasskey represents the relationship between a user and their passkey credentials
type UserPasskey struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	CredentialID primitive.ObjectID `bson:"credentialId" json:"credentialId"`
	Name         string             `bson:"name" json:"name"`
	DeviceInfo   string             `bson:"deviceInfo" json:"deviceInfo"`
	IsActive     bool               `bson:"isActive" json:"isActive"`
	LastUsedAt   time.Time          `bson:"lastUsedAt" json:"lastUsedAt"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

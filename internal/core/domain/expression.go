package domain

import (
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaContent holds the temporary upload data
type MediaContent struct {
	Reader   io.Reader
	Filename string
}

type Expression struct {
	ID                         primitive.ObjectID       `bson:"_id,omitempty"`
	Creator                    string                   `bson:"creator"`
	CreatorAddress             string                   `bson:"creatorAddress"`
	Content                    map[string]string        `bson:"content"` // Stored content paths
	MediaContent               map[string]*MediaContent `bson:"-"`       // Temporary media upload data
	IPFSHash                   string                   `bson:"ipfsHash"`
	OnChainID                  int                      `bson:"onChainId"`
	Status                     string                   `bson:"status"`
	Acknowledgements           []*Acknowledgement       `bson:"-"`
	IsAcknowledged             bool                     `bson:"-"`
	ActiveAcknowledgementCount int                      `bson:"-"`
	CreatedAt                  time.Time                `bson:"createdAt"`
	UpdatedAt                  time.Time                `bson:"updatedAt"`
}

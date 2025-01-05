package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Statistics represents a snapshot of system metrics
type Statistics struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TotalExpressions      int                `bson:"totalExpressions" json:"totalExpressions"`
	TotalAcknowledgements int                `bson:"totalAcknowledgements" json:"totalAcknowledgements"`
	TotalUsers            int                `bson:"totalUsers" json:"totalUsers"`
	CitizenshipStats      map[string]int     `bson:"citizenshipStats" json:"citizenshipStats"` // country code -> count
	MediaStats            map[string]int     `bson:"mediaStats" json:"mediaStats"`             // media type -> count
	CreatedAt             time.Time          `bson:"createdAt" json:"createdAt"`
}

// MediaType constants for tracking different types of media in expressions
const (
	MediaTypeText  = "text"
	MediaTypeImage = "image"
	MediaTypeAudio = "audio"
	MediaTypeVideo = "video"
)

// CountryInfo represents a country's information for the UI
type CountryInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Flag string `json:"flag"` // URL or emoji for the flag
}

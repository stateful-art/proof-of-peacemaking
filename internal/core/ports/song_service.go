package ports

import (
	"context"
	"proofofpeacemaking/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SongService interface {
	AddSong(ctx context.Context, url string, userAddress string) error
	GetNextSong(ctx context.Context) (*domain.Song, error)
	MarkSongAsPlayed(ctx context.Context, songID primitive.ObjectID) error
	GetCurrentlyPlaying(ctx context.Context) (*domain.Song, error)
	GetQueue(ctx context.Context) ([]*domain.Song, error)
	GetArchive(ctx context.Context) ([]*domain.Song, error)
}

type SongRepository interface {
	AddSong(ctx context.Context, song *domain.Song) error
	GetUnplayedSongs(ctx context.Context) ([]*domain.Song, error)
	MarkAsPlaying(ctx context.Context, songID primitive.ObjectID) error
	MarkAsPlayed(ctx context.Context, songID primitive.ObjectID) error
	GetCurrentlyPlaying(ctx context.Context) (*domain.Song, error)
	GetQueue(ctx context.Context) ([]*domain.Song, error)
	GetArchive(ctx context.Context) ([]*domain.Song, error)
	EnsureIndexes() error
}

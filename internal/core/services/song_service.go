package services

import (
	"context"
	"errors"
	"math/rand"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SongService struct {
	repo ports.SongRepository
}

func NewSongService(repo ports.SongRepository) *SongService {
	return &SongService{
		repo: repo,
	}
}

func (s *SongService) AddSong(ctx context.Context, url string, userAddress string) error {
	// Validate URL format (should be a single video, not a playlist)
	if strings.Contains(url, "&list=") || !strings.Contains(url, "youtube.com/watch?v=") {
		return errors.New("invalid URL format: must be a single YouTube video")
	}

	song := &domain.Song{
		URL:     url,
		AddedBy: userAddress,
	}

	return s.repo.AddSong(ctx, song)
}

func (s *SongService) GetNextSong(ctx context.Context) (*domain.Song, error) {
	// First check if any song is currently playing
	currentSong, err := s.repo.GetCurrentlyPlaying(ctx)
	if err != nil {
		return nil, err
	}

	// If there's a song playing, don't start another one
	if currentSong != nil {
		return currentSong, nil
	}

	// Get all unplayed songs
	songs, err := s.repo.GetUnplayedSongs(ctx)
	if err != nil {
		return nil, err
	}

	if len(songs) == 0 {
		return nil, nil
	}

	// Randomly select a song
	rand.Seed(time.Now().UnixNano())
	selectedSong := songs[rand.Intn(len(songs))]

	// Mark it as playing
	err = s.repo.MarkAsPlaying(ctx, selectedSong.ID)
	if err != nil {
		return nil, err
	}

	return selectedSong, nil
}

func (s *SongService) MarkSongAsPlayed(ctx context.Context, songID primitive.ObjectID) error {
	// First mark the current song as played
	err := s.repo.MarkAsPlayed(ctx, songID)
	if err != nil {
		return err
	}

	// Then try to get and start the next song immediately
	nextSong, err := s.GetNextSong(ctx)
	if err != nil {
		return err
	}

	// If there's no next song, that's okay - just return
	if nextSong == nil {
		return nil
	}

	return nil
}

func (s *SongService) GetCurrentlyPlaying(ctx context.Context) (*domain.Song, error) {
	return s.repo.GetCurrentlyPlaying(ctx)
}

func (s *SongService) GetQueue(ctx context.Context) ([]*domain.Song, error) {
	return s.repo.GetQueue(ctx)
}

func (s *SongService) GetArchive(ctx context.Context) ([]*domain.Song, error) {
	return s.repo.GetArchive(ctx)
}

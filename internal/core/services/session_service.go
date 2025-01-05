package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"
)

type sessionService struct {
	sessionRepo ports.SessionRepository
}

func NewSessionService(sessionRepo ports.SessionRepository) ports.SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) Create(ctx context.Context, session *domain.Session) error {
	// Generate a random token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Set session fields
	session.Token = token
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(24 * time.Hour) // 24-hour expiry

	return s.sessionRepo.Create(ctx, session)
}

func (s *sessionService) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has expired
	if session != nil && time.Now().After(session.ExpiresAt) {
		if err := s.Delete(ctx, token); err != nil {
			return nil, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return nil, nil
	}

	return session, nil
}

func (s *sessionService) Update(ctx context.Context, session *domain.Session) error {
	session.UpdatedAt = time.Now()
	return s.sessionRepo.Update(ctx, session)
}

func (s *sessionService) Delete(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, token)
}

// generateToken generates a random token for session identification
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

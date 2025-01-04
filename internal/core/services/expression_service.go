package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"proofofpeacemaking/internal/core/storage"
	"time"
)

type expressionService struct {
	expressionRepo      ports.ExpressionRepository
	acknowledgementRepo ports.AcknowledgementRepository
	storage             storage.Storage
}

func NewExpressionService(
	expressionRepo ports.ExpressionRepository,
	acknowledgementRepo ports.AcknowledgementRepository,
	storage storage.Storage,
) ports.ExpressionService {
	return &expressionService{
		expressionRepo:      expressionRepo,
		acknowledgementRepo: acknowledgementRepo,
		storage:             storage,
	}
}

func (s *expressionService) Create(ctx context.Context, expression *domain.Expression) error {
	// Handle media uploads if present
	if expression.MediaContent != nil {
		// Handle video upload
		if media, ok := expression.MediaContent["video"]; ok {
			if key, err := s.UploadMedia(ctx, expression.ID.Hex(), "video", media.Reader, media.Filename); err == nil {
				expression.Content["video"] = key
			}
		}

		// Handle audio upload
		if media, ok := expression.MediaContent["audio"]; ok {
			if key, err := s.UploadMedia(ctx, expression.ID.Hex(), "audio", media.Reader, media.Filename); err == nil {
				expression.Content["audio"] = key
			}
		}

		// Handle image upload
		if media, ok := expression.MediaContent["image"]; ok {
			if key, err := s.UploadMedia(ctx, expression.ID.Hex(), "image", media.Reader, media.Filename); err == nil {
				expression.Content["image"] = key
			}
		}

		// Clear temporary media content
		expression.MediaContent = nil
	}

	// Create expression in repository with updated content paths
	if err := s.expressionRepo.Create(ctx, expression); err != nil {
		return fmt.Errorf("failed to create expression: %w", err)
	}
	return nil
}

func (s *expressionService) Get(ctx context.Context, id string) (*domain.Expression, error) {
	expression, err := s.expressionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expression: %w", err)
	}
	if expression == nil {
		return nil, nil
	}

	// Initialize counts to 0
	expression.ActiveAcknowledgementCount = 0
	expression.Acknowledgements = []*domain.Acknowledgement{}

	// Get acknowledgements for the expression
	acks, err := s.acknowledgementRepo.FindByExpression(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
	}
	expression.Acknowledgements = acks

	// Calculate active acknowledgement count
	for _, ack := range acks {
		if ack.Status == domain.AcknowledgementStatusActive {
			expression.ActiveAcknowledgementCount++
		}
	}

	// Generate presigned URLs for media content
	if err := s.addPresignedURLs(ctx, expression); err != nil {
		return nil, fmt.Errorf("failed to generate presigned URLs: %w", err)
	}

	return expression, nil
}

func (s *expressionService) List(ctx context.Context) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list expressions: %w", err)
	}

	// For each expression, get its acknowledgements and calculate counts
	for _, expr := range expressions {
		// Initialize counts to 0
		expr.ActiveAcknowledgementCount = 0
		expr.Acknowledgements = []*domain.Acknowledgement{}

		// Get acknowledgements
		acks, err := s.acknowledgementRepo.FindByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
		}
		expr.Acknowledgements = acks

		// Calculate active acknowledgement count
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				expr.ActiveAcknowledgementCount++
			}
		}

		// Generate presigned URLs for media content
		if err := s.addPresignedURLs(ctx, expr); err != nil {
			return nil, fmt.Errorf("failed to generate presigned URLs: %w", err)
		}
	}

	return expressions, nil
}

func (s *expressionService) ListByUser(ctx context.Context, userID string) ([]*domain.Expression, error) {
	expressions, err := s.expressionRepo.FindByCreatorID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expressions by user: %w", err)
	}

	// For each expression, get its acknowledgements and calculate counts
	for _, expr := range expressions {
		// Initialize counts to 0
		expr.ActiveAcknowledgementCount = 0
		expr.Acknowledgements = []*domain.Acknowledgement{}

		// Get acknowledgements
		acks, err := s.acknowledgementRepo.FindByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
		}
		expr.Acknowledgements = acks

		// Calculate active acknowledgement count
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				expr.ActiveAcknowledgementCount++
			}
		}
	}

	return expressions, nil
}

func (s *expressionService) GetMultiple(ctx context.Context, ids []string) (map[string]*domain.Expression, error) {
	expressions := make(map[string]*domain.Expression)

	// Get all expressions in one query
	expressionsList, err := s.expressionRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get expressions: %w", err)
	}

	// Create a map of expression IDs to expressions
	for _, expr := range expressionsList {
		// Initialize counts to 0
		expr.ActiveAcknowledgementCount = 0
		expr.Acknowledgements = []*domain.Acknowledgement{}

		// Get acknowledgements for the expression
		acks, err := s.acknowledgementRepo.FindByExpression(ctx, expr.ID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get acknowledgements for expression: %w", err)
		}
		expr.Acknowledgements = acks

		// Calculate active acknowledgement count
		for _, ack := range acks {
			if ack.Status == domain.AcknowledgementStatusActive {
				expr.ActiveAcknowledgementCount++
			}
		}

		expressions[expr.ID.Hex()] = expr
	}

	return expressions, nil
}

// Helper function to get content type
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	default:
		return "application/octet-stream"
	}
}

// UploadMedia uploads media content for an expression
func (s *expressionService) UploadMedia(ctx context.Context, expressionID string, mediaType string, reader io.Reader, filename string) (string, error) {
	log.Printf("[EXPRESSION] Starting media upload - ExpressionID: %s, MediaType: %s, Filename: %s", expressionID, mediaType, filename)

	// Get the file extension from original filename
	ext := filepath.Ext(filename)
	// Path format: expressions/[expressionID]/[mediaType][extension]
	// Example: expressions/123abc/video.mp4
	key := fmt.Sprintf("expressions/%s/%s%s", expressionID, mediaType, ext)
	log.Printf("[EXPRESSION] Generated storage key: %s", key)

	// Use the content type detection
	contentType := getContentType(filename)
	log.Printf("[EXPRESSION] Detected content type: %s for file: %s", contentType, filename)

	err := s.storage.UploadFile(ctx, key, reader, storage.UploadOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=31536000", // 1 year cache for media
	})
	if err != nil {
		log.Printf("[EXPRESSION] Failed to upload media - Key: %s, Error: %v", key, err)
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	log.Printf("[EXPRESSION] Successfully uploaded media - Key: %s", key)
	return key, nil
}

// GetMedia retrieves media content for an expression
func (s *expressionService) GetMedia(ctx context.Context, expressionID string, mediaType string) (io.ReadCloser, error) {
	key := fmt.Sprintf("expressions/%s/%s", expressionID, mediaType)
	log.Printf("[EXPRESSION] Retrieving media - ExpressionID: %s, MediaType: %s, Key: %s", expressionID, mediaType, key)

	reader, err := s.storage.GetFile(ctx, key)
	if err != nil {
		log.Printf("[EXPRESSION] Failed to retrieve media - Key: %s, Error: %v", key, err)
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	log.Printf("[EXPRESSION] Successfully retrieved media - Key: %s", key)
	return reader, nil
}

// DeleteMedia removes media content for an expression
func (s *expressionService) DeleteMedia(ctx context.Context, expressionID string, mediaType string) error {
	key := fmt.Sprintf("expressions/%s/%s", expressionID, mediaType)
	log.Printf("[EXPRESSION] Deleting media - ExpressionID: %s, MediaType: %s, Key: %s", expressionID, mediaType, key)

	if err := s.storage.DeleteFile(ctx, key); err != nil {
		log.Printf("[EXPRESSION] Failed to delete media - Key: %s, Error: %v", key, err)
		return fmt.Errorf("failed to delete media: %w", err)
	}

	log.Printf("[EXPRESSION] Successfully deleted media - Key: %s", key)
	return nil
}

// Helper function to add presigned URLs to an expression's content
func (s *expressionService) addPresignedURLs(ctx context.Context, expression *domain.Expression) error {
	log.Printf("[EXPRESSION] Generating presigned URLs for expression: %s", expression.ID.Hex())

	mediaTypes := []string{"image", "audio", "video"}
	for _, mediaType := range mediaTypes {
		if key, exists := expression.Content[mediaType]; exists {
			log.Printf("[EXPRESSION] Generating presigned URL for %s - Key: %s", mediaType, key)

			// Generate a presigned URL that's valid for 1 hour
			url, err := s.storage.GetPresignedURL(ctx, key, time.Hour)
			if err != nil {
				log.Printf("[EXPRESSION] Failed to generate presigned URL for %s - Key: %s, Error: %v", mediaType, key, err)
				return fmt.Errorf("failed to generate presigned URL for %s: %w", mediaType, err)
			}

			expression.Content[mediaType] = url
			log.Printf("[EXPRESSION] Successfully generated presigned URL for %s", mediaType)
		}
	}

	return nil
}

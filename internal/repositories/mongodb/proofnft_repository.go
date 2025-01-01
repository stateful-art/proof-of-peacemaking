package mongodb

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type proofNFTRepository struct {
	collection *mongo.Collection
}

func NewProofNFTRepository(db *mongo.Database) ports.ProofNFTRepository {
	return &proofNFTRepository{
		collection: db.Collection("proofnfts"),
	}
}

func (r *proofNFTRepository) Create(ctx context.Context, proofNFT *domain.ProofNFT) error {
	_, err := r.collection.InsertOne(ctx, proofNFT)
	if err != nil {
		return fmt.Errorf("failed to insert proof NFT: %w", err)
	}
	return nil
}

func (r *proofNFTRepository) Update(ctx context.Context, proofNFT *domain.ProofNFT) error {
	filter := bson.M{"_id": proofNFT.ID}
	update := bson.M{"$set": proofNFT}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update proof NFT: %w", err)
	}
	return nil
}

func (r *proofNFTRepository) FindByID(ctx context.Context, id string) (*domain.ProofNFT, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	var proofNFT domain.ProofNFT
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&proofNFT)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find proof NFT: %w", err)
	}
	return &proofNFT, nil
}

func (r *proofNFTRepository) FindByAcknowledger(ctx context.Context, acknowledgerID string) ([]*domain.ProofNFT, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"acknowledgerId": acknowledgerID})
	if err != nil {
		return nil, fmt.Errorf("failed to find proof NFTs: %w", err)
	}
	defer cursor.Close(ctx)

	var proofNFTs []*domain.ProofNFT
	if err := cursor.All(ctx, &proofNFTs); err != nil {
		return nil, fmt.Errorf("failed to decode proof NFTs: %w", err)
	}
	return proofNFTs, nil
}

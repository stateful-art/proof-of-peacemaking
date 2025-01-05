package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect() *mongo.Database {
	log.Printf("[DB] Connecting to MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGODB_URI")
	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("[DB] Failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("[DB] Failed to ping MongoDB: %v", err)
	}

	log.Printf("[DB] Successfully connected to MongoDB")
	db := client.Database("proof-of-peacemaking")

	// Drop and recreate indexes
	if err := dropAndRecreateIndexes(ctx, db); err != nil {
		log.Fatalf("[DB] Failed to recreate indexes: %v", err)
	}

	return db
}

func dropAndRecreateIndexes(ctx context.Context, db *mongo.Database) error {
	// Drop all indexes from users collection
	log.Printf("[DB] Dropping all indexes from users collection...")
	_, err := db.Collection("users").Indexes().DropAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop indexes: %w", err)
	}

	// Recreate indexes
	log.Printf("[DB] Recreating indexes...")
	if err := createIndexes(ctx, db); err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	return nil
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	configs := []IndexConfig{
		{
			Collection: "users",
			Fields: []IndexField{
				{Name: "address", Order: 1, Unique: true, Sparse: true},
				{Name: "email", Order: 1, Unique: true, Sparse: true},
				{Name: "username", Order: 1, Unique: true, Sparse: true},
			},
		},
		{
			Collection: "expressions",
			Fields: []IndexField{
				{Name: "creator", Order: 1},
				{Name: "createdAt", Order: -1},
			},
		},
		{
			Collection: "acknowledgements",
			Fields: []IndexField{
				{Name: "expressionId", Order: 1, Compound: true},
				{Name: "acknowledger", Order: 1, Compound: true},
			},
		},
	}

	return EnsureIndexes(ctx, db, configs)
}

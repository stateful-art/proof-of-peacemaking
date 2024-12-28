package mongodb

import (
	"context"
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

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatalf("[DB] Failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("[DB] Failed to ping MongoDB: %v", err)
	}

	log.Printf("[DB] Successfully connected to MongoDB")
	db := client.Database("proof-of-peacemaking")

	// Create indexes if needed
	if err := createIndexes(ctx, db); err != nil {
		log.Fatalf("[DB] Failed to create indexes: %v", err)
	}

	return db
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Create unique index on user address
	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{
			"address": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Printf("[DB] Created unique index on users.address")
	return nil
}

package mongodb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexConfig struct {
	Collection string
	Fields     []IndexField
}

type IndexField struct {
	Name     string
	Order    int // 1 for ascending, -1 for descending
	Unique   bool
	Compound bool // true if this field is part of a compound index
}

func EnsureIndexes(ctx context.Context, db *mongo.Database, configs []IndexConfig) error {
	for _, config := range configs {
		indexView := db.Collection(config.Collection).Indexes()

		// Group compound indexes
		compoundFields := make(map[string]bson.D)
		singleFields := []IndexField{}

		for _, field := range config.Fields {
			if field.Compound {
				key := "compound" // You might want to add a group identifier for multiple compound indexes
				compoundFields[key] = append(compoundFields[key], bson.E{Key: field.Name, Value: field.Order})
			} else {
				singleFields = append(singleFields, field)
			}
		}

		// Create compound indexes
		for _, fields := range compoundFields {
			if err := createIndex(ctx, indexView, fields, false, config.Collection); err != nil {
				return err
			}
		}

		// Create single field indexes
		for _, field := range singleFields {
			keys := bson.D{{Key: field.Name, Value: field.Order}}
			if err := createIndex(ctx, indexView, keys, field.Unique, config.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func createIndex(ctx context.Context, indexView mongo.IndexView, keys bson.D, unique bool, collection string) error {
	// Check if index already exists
	cursor, err := indexView.List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err = cursor.All(ctx, &indexes); err != nil {
		return err
	}

	// Generate index name
	indexName := generateIndexName(keys)

	// Check if index exists
	for _, index := range indexes {
		if index["name"] == indexName {
			return nil
		}
	}

	// Create index
	_, err = indexView.CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(unique),
	})
	if err != nil {
		return err
	}

	// Improved log message
	keyInfo := ""
	for i, key := range keys {
		if i > 0 {
			keyInfo += ", "
		}
		order := "ASC"
		if key.Value.(int) == -1 {
			order = "DESC"
		}
		keyInfo += fmt.Sprintf("%s:%s", key.Key, order)
	}

	log.Printf("[DB] Created new index '%s' on collection '%s' (%s)", indexName, collection, keyInfo)
	return nil
}

func generateIndexName(keys bson.D) string {
	name := ""
	for _, key := range keys {
		if name != "" {
			name += "_"
		}
		name += key.Key + "_" + "1"
	}
	return name
}

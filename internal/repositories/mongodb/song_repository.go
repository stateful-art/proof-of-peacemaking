package mongodb

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type songRepository struct {
	collection *mongo.Collection
}

func NewSongRepository(db *mongo.Database) ports.SongRepository {
	collection := db.Collection("songs")
	return &songRepository{
		collection: collection,
	}
}

func (r *songRepository) EnsureIndexes() error {
	_, err := r.collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "url", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *songRepository) AddSong(ctx context.Context, song *domain.Song) error {
	song.AddedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, song)
	return err
}

func (r *songRepository) GetUnplayedSongs(ctx context.Context) ([]*domain.Song, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"played_at":  nil,
		"is_playing": false,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*domain.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *songRepository) MarkAsPlaying(ctx context.Context, songID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": songID},
		bson.M{
			"$set": bson.M{
				"is_playing": true,
			},
		},
	)
	return err
}

func (r *songRepository) MarkAsPlayed(ctx context.Context, songID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": songID},
		bson.M{
			"$set": bson.M{
				"played_at":  time.Now(),
				"is_playing": false,
			},
		},
	)
	return err
}

func (r *songRepository) GetCurrentlyPlaying(ctx context.Context) (*domain.Song, error) {
	var song domain.Song
	err := r.collection.FindOne(ctx, bson.M{"is_playing": true}).Decode(&song)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &song, err
}

func (r *songRepository) GetQueue(ctx context.Context) ([]*domain.Song, error) {
	opts := options.Find().SetSort(bson.D{{Key: "added_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{
		"played_at": nil,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*domain.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *songRepository) GetArchive(ctx context.Context) ([]*domain.Song, error) {
	opts := options.Find().SetSort(bson.D{{Key: "played_at", Value: -1}}).SetLimit(50)
	cursor, err := r.collection.Find(ctx, bson.M{
		"played_at": bson.M{"$ne": nil},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*domain.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}

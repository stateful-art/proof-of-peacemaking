package mongodb

import (
	"context"
	"time"

	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type notificationRepository struct {
	db *mongo.Database
}

func NewNotificationRepository(db *mongo.Database) ports.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification domain.Notification) error {
	_, err := r.db.Collection("notifications").InsertOne(ctx, notification)
	return err
}

func (r *notificationRepository) GetByUser(ctx context.Context, userID string) ([]domain.Notification, error) {
	cursor, err := r.db.Collection("notifications").Find(ctx, bson.M{
		"userId": userID,
	})
	if err != nil {
		return nil, err
	}

	var notifications []domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, userID string, notificationID string) error {
	_, err := r.db.Collection("notifications").UpdateOne(
		ctx,
		bson.M{
			"userId": userID,
			"_id":    notificationID,
		},
		bson.M{
			"$set": bson.M{
				"read":   true,
				"readAt": time.Now(),
			},
		},
	)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := r.db.Collection("notifications").UpdateMany(
		ctx,
		bson.M{
			"userId": userID,
			"read":   false,
		},
		bson.M{
			"$set": bson.M{
				"read":   true,
				"readAt": time.Now(),
			},
		},
	)
	return err
}

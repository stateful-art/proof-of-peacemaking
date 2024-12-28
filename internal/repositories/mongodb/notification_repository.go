package mongodb

import (
	"context"

	"github.com/stateful-art/proof-of-peacemaking/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type notificationRepository struct {
	db *mongo.Database
}

func NewNotificationRepository(db *mongo.Database) *notificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	_, err := r.db.Collection("notifications").InsertOne(ctx, notification)
	return err
}

func (r *notificationRepository) CreateUserNotification(ctx context.Context, un *domain.UserNotification) error {
	_, err := r.db.Collection("user_notifications").InsertOne(ctx, un)
	return err
}

func (r *notificationRepository) GetUserUnreadNotifications(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error) {
	pipeline := mongo.Pipeline{
		{{
			"$match": bson.M{
				"userId": userID,
				"read":   false,
			},
		}},
		{{
			"$lookup": bson.M{
				"from":         "notifications",
				"localField":   "notificationId",
				"foreignField": "_id",
				"as":           "notification",
			},
		}},
		{{"$unwind": "$notification"}},
	}

	cursor, err := r.db.Collection("user_notifications").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var notifications []*domain.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

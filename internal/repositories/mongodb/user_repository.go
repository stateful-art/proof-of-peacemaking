package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"proofofpeacemaking/internal/core/domain"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	// Check if email already exists (case-insensitive)
	if user.Email != "" {
		exists, err := r.emailExists(ctx, user.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already exists")
		}
	}

	// Check if username already exists (case-insensitive)
	if user.Username != "" {
		exists, err := r.usernameExists(ctx, user.Username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username already exists")
		}
	}

	// Check if address already exists (if provided)
	if user.Address != "" {
		exists, err := r.addressExists(ctx, user.Address)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("address already exists")
		}
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	updateFields := bson.M{"updatedAt": time.Now()}

	// Only check and update email if it's being changed
	if user.Email != "" {
		exists, err := r.emailExistsForOtherUser(ctx, user.Email, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already exists")
		}
		updateFields["email"] = user.Email
	}

	// Only check and update username if it's being changed
	if user.Username != "" {
		exists, err := r.usernameExistsForOtherUser(ctx, user.Username, user.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username already exists")
		}
		updateFields["username"] = user.Username
	}

	// Update other non-unique fields if they're set
	if user.DisplayName != "" {
		updateFields["displayName"] = user.DisplayName
	}
	if user.Citizenship != "" {
		updateFields["citizenship"] = user.Citizenship
	}
	if user.City != "" {
		updateFields["city"] = user.City
	}
	if user.Password != "" {
		updateFields["password"] = user.Password
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": updateFields}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = r.db.Collection("users").FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByAddress(ctx context.Context, address string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"email": bson.M{"$regex": "^" + regexp.QuoteMeta(email) + "$", "$options": "i"},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) emailExistsForOtherUser(ctx context.Context, email string, userID primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"email": bson.M{"$regex": "^" + regexp.QuoteMeta(email) + "$", "$options": "i"},
		"_id":   bson.M{"$ne": userID},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) usernameExists(ctx context.Context, username string) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"username": bson.M{"$regex": "^" + regexp.QuoteMeta(username) + "$", "$options": "i"},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) usernameExistsForOtherUser(ctx context.Context, username string, userID primitive.ObjectID) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"username": bson.M{"$regex": "^" + regexp.QuoteMeta(username) + "$", "$options": "i"},
		"_id":      bson.M{"$ne": userID},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error {
	// Check if wallet is already connected to another user
	exists, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"address": address,
		"_id":     bson.M{"$ne": userID},
	})
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("wallet already connected to another account")
	}

	// Update user with wallet address
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"address":   address,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"nonce":     nonce,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.db.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update nonce: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.db.Collection("users").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// GetCitizenshipDistribution returns a map of citizenship counts
func (r *UserRepository) GetCitizenshipDistribution(ctx context.Context) (map[string]int, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$citizenship",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	for _, result := range results {
		if result.ID != "" { // Skip empty citizenship
			distribution[result.ID] = result.Count
		}
	}

	return distribution, nil
}

// GetTotalCount returns the total number of users in the system
func (r *UserRepository) GetTotalCount(ctx context.Context) (int, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to get total user count: %w", err)
	}
	return int(count), nil
}

func (r *UserRepository) cleanupDuplicates(ctx context.Context) error {
	// Find all users
	cursor, err := r.db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	defer cursor.Close(ctx)

	// Map to track seen emails and usernames
	seenEmails := make(map[string]primitive.ObjectID)
	seenUsernames := make(map[string]primitive.ObjectID)

	// Process each user
	for cursor.Next(ctx) {
		var user domain.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("[WARN] Failed to decode user: %v", err)
			continue
		}

		// Handle duplicate emails
		if user.Email != "" {
			emailLower := strings.ToLower(user.Email)
			if existingID, exists := seenEmails[emailLower]; exists {
				// Keep the older record (smaller ObjectID), clear email from newer one
				if user.ID.Hex() > existingID.Hex() {
					log.Printf("[INFO] Clearing duplicate email %s from user %s", user.Email, user.ID.Hex())
					if err := r.clearEmail(ctx, user.ID); err != nil {
						log.Printf("[ERROR] Failed to clear email: %v", err)
					}
				} else {
					log.Printf("[INFO] Clearing duplicate email %s from user %s", user.Email, existingID.Hex())
					if err := r.clearEmail(ctx, existingID); err != nil {
						log.Printf("[ERROR] Failed to clear email: %v", err)
					}
					seenEmails[emailLower] = user.ID
				}
			} else {
				seenEmails[emailLower] = user.ID
			}
		}

		// Handle duplicate usernames
		if user.Username != "" {
			usernameLower := strings.ToLower(user.Username)
			if existingID, exists := seenUsernames[usernameLower]; exists {
				// Keep the older record (smaller ObjectID), clear username from newer one
				if user.ID.Hex() > existingID.Hex() {
					log.Printf("[INFO] Clearing duplicate username %s from user %s", user.Username, user.ID.Hex())
					if err := r.clearUsername(ctx, user.ID); err != nil {
						log.Printf("[ERROR] Failed to clear username: %v", err)
					}
				} else {
					log.Printf("[INFO] Clearing duplicate username %s from user %s", user.Username, existingID.Hex())
					if err := r.clearUsername(ctx, existingID); err != nil {
						log.Printf("[ERROR] Failed to clear username: %v", err)
					}
					seenUsernames[usernameLower] = user.ID
				}
			} else {
				seenUsernames[usernameLower] = user.ID
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

func (r *UserRepository) clearEmail(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"email": ""}},
	)
	return err
}

func (r *UserRepository) clearUsername(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"username": ""}},
	)
	return err
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	log.Printf("[DB] Dropping all indexes from users collection...")
	_, err := r.db.Collection("users").Indexes().DropAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop indexes: %w", err)
	}

	log.Printf("[DB] Cleaning up duplicate entries...")
	if err := r.cleanupDuplicates(ctx); err != nil {
		return fmt.Errorf("failed to cleanup duplicates: %w", err)
	}

	log.Printf("[DB] Recreating indexes...")
	// Create a unique index for non-null addresses
	addressIndex := mongo.IndexModel{
		Keys:    bson.D{{"address", 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}

	// Create a unique index for non-empty emails (case-insensitive)
	emailIndex := mongo.IndexModel{
		Keys: bson.D{{"email", 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case-insensitive
		}),
	}

	// Create a unique index for non-empty usernames (case-insensitive)
	usernameIndex := mongo.IndexModel{
		Keys: bson.D{{"username", 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case-insensitive
		}),
	}

	collection := r.db.Collection("users")
	if _, err := collection.Indexes().CreateOne(ctx, addressIndex); err != nil {
		return fmt.Errorf("failed to create address index: %w", err)
	}
	log.Printf("[DB] Created new index 'address_1' on collection 'users' (address:ASC)")

	if _, err := collection.Indexes().CreateOne(ctx, emailIndex); err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}
	log.Printf("[DB] Created new index 'email_1' on collection 'users' (email:ASC)")

	if _, err := collection.Indexes().CreateOne(ctx, usernameIndex); err != nil {
		return fmt.Errorf("failed to create username index: %w", err)
	}
	log.Printf("[DB] Created new index 'username_1' on collection 'users' (username:ASC)")

	return nil
}

func (r *UserRepository) addressExists(ctx context.Context, address string) (bool, error) {
	count, err := r.db.Collection("users").CountDocuments(ctx, bson.M{
		"address": address,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/7-solutions/backend-challenge/internal/domain/model"
	"github.com/7-solutions/backend-challenge/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Common errors
var (
	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrDatabase       = errors.New("database error")
)

// mongoRepository implements the UserRepository interface
type mongoRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(ctx context.Context, uri, dbName string) (repository.UserRepository, error) {
	// Set up connection options
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	// Create repository
	repo := &mongoRepository{
		client:     client,
		database:   dbName,
		collection: "users",
	}

	// Create indexes for email uniqueness
	err = repo.createIndexes(ctx)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Create unique index for email
func (r *mongoRepository) createIndexes(ctx context.Context) error {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	// Create unique index for email
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	
	return err
}

// Create adds a new user to the database
func (r *mongoRepository) Create(ctx context.Context, user *model.User) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Generate new ID if not set
	if user.ID == "" {
		user.ID = primitive.NewObjectID().Hex()
	}

	// Set creation time if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	
	// Insert the document
	_, err := collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateEmail
	}
	
	return err
}

// GetByID fetches a user by ID
func (r *mongoRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	// Check if ID is a valid ObjectID
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		// Try to find by ObjectID or string ID
		filter = bson.M{
			"$or": []bson.M{
				{"_id": objectID},
				{"_id": id},
			},
		}
	} else {
		// Use string ID
		filter = bson.M{"_id": id}
	}

	var user model.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// GetByEmail fetches a user by email
func (r *mongoRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	var user model.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// Update updates a user in the database
func (r *mongoRepository) Update(ctx context.Context, user *model.User) error {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	// Check if ID is a valid ObjectID
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(user.ID); err == nil {
		// Try to update by ObjectID or string ID
		filter = bson.M{
			"$or": []bson.M{
				{"_id": objectID},
				{"_id": user.ID},
			},
		}
	} else {
		// Use string ID
		filter = bson.M{"_id": user.ID}
	}
	
	update := bson.M{
		"$set": bson.M{
			"name":  user.Name,
			"email": user.Email,
		},
	}
	
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

// Delete removes a user from the database
func (r *mongoRepository) Delete(ctx context.Context, id string) error {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	// Check if ID is a valid ObjectID
	var filter bson.M
	if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
		// Try to delete by ObjectID or string ID
		filter = bson.M{
			"$or": []bson.M{
				{"_id": objectID},
				{"_id": id},
			},
		}
	} else {
		// Use string ID
		filter = bson.M{"_id": id}
	}
	
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}
	
	return nil
}

// List returns all users with pagination
func (r *mongoRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	// Calculate skip for pagination
	skip := (page - 1) * pageSize
	
	// Set up find options
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at desc
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))
	
	// Perform find
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	// Decode results
	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	
	return users, nil
}

// CountUsers returns the total number of users in the database
func (r *mongoRepository) CountUsers(ctx context.Context) (int64, error) {
	collection := r.client.Database(r.database).Collection(r.collection)
	
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// Disconnect closes the MongoDB connection
func (r *mongoRepository) Disconnect(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"
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

// MongoRepository implements the UserRepository interface
type MongoRepository struct {
	Client     *mongo.Client
	database   string
	collection string
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(ctx context.Context, uri, dbName string) (repository.UserRepository, error) {
	// Use MongoDB in production, fall back to mock for development or testing
	if uri == "" {
		return nil, errors.New("MongoDB URI is not provided")
	}
	
	if uri == "mock" {
		log.Println("WARNING: Using in-memory mock repository as explicitly requested")
		return NewMockRepository(), nil
	}

	// Set client options with reasonable timeouts
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(5 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Set up a timeout context for the connection test
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Check the connection
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		// Disconnect the client on failure
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB successfully")

	// Create repository
	repo := &MongoRepository{
		Client:     client,
		database:   dbName,
		collection: "users",
	}

	// Create indexes
	err = repo.createIndexes(ctx)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// mockRepository implements the UserRepository interface with in-memory storage
type mockRepository struct {
	users map[string]*model.User
	mu    sync.RWMutex
}

// NewMockRepository creates a new mock repository
func NewMockRepository() repository.UserRepository {
	return &mockRepository{
		users: make(map[string]*model.User),
	}
}

// Create adds a new user
func (r *mockRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate email
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return ErrDuplicateEmail
		}
	}

	// Generate ID if not set
	if user.ID == "" {
		user.ID = primitive.NewObjectID().Hex()
	}

	// Set creation time if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// Store user
	r.users[user.ID] = user
	return nil
}

// GetByID fetches a user by ID
func (r *mockRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByEmail fetches a user by email
func (r *mockRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// Update updates a user
func (r *mockRepository) Update(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	if _, ok := r.users[user.ID]; !ok {
		return ErrUserNotFound
	}

	// Check for duplicate email
	for id, existingUser := range r.users {
		if id != user.ID && existingUser.Email == user.Email {
			return ErrDuplicateEmail
		}
	}

	// Update user
	r.users[user.ID] = user
	return nil
}

// Delete removes a user
func (r *mockRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

// List returns all users with pagination
func (r *mockRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*model.User
	for _, user := range r.users {
		users = append(users, user)
	}

	// Sort by created_at
	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.After(users[j].CreatedAt)
	})

	// Paginate
	start := (page - 1) * pageSize
	if start >= len(users) {
		return []*model.User{}, nil
	}

	end := start + pageSize
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

// CountUsers returns the total number of users
func (r *mockRepository) CountUsers(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.users)), nil
}

// Disconnect closes the connection
func (r *mockRepository) Disconnect(ctx context.Context) error {
	return nil
}

// Create unique index for email
func (r *MongoRepository) createIndexes(ctx context.Context) error {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) Create(ctx context.Context, user *model.User) error {
	collection := r.Client.Database(r.database).Collection(r.collection)

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
func (r *MongoRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) Update(ctx context.Context, user *model.User) error {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
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
func (r *MongoRepository) CountUsers(ctx context.Context) (int64, error) {
	collection := r.Client.Database(r.database).Collection(r.collection)
	
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// Disconnect closes the MongoDB connection
func (r *MongoRepository) Disconnect(ctx context.Context) error {
	return r.Client.Disconnect(ctx)
}
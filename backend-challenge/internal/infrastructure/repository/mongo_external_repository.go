package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoExternalRepository implements the ExternalUserRepository interface
type mongoExternalRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoExternalRepository creates a new MongoDB repository for external users
func NewMongoExternalRepository(ctx context.Context, client *mongo.Client, dbName string) repository.ExternalUserRepository {
	repo := &mongoExternalRepository{
		client:     client,
		database:   dbName,
		collection: "external_users",
	}

	// Create index for better querying
	repo.createIndexes(ctx)

	return repo
}

// Create index for better querying
func (r *mongoExternalRepository) createIndexes(ctx context.Context) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Index on department for faster department-based queries
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "company.department", Value: 1}},
		},
	)

	return err
}

// Create adds a new external user
func (r *mongoExternalRepository) Create(ctx context.Context, user *model.ExternalUser) error {
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
	return err
}

// GetByID fetches an external user by ID
func (r *mongoExternalRepository) GetByID(ctx context.Context, id string) (*model.ExternalUser, error) {
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

	var user model.ExternalUser
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update updates an external user
func (r *mongoExternalRepository) Update(ctx context.Context, user *model.ExternalUser) error {
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

	// Update user
	result, err := collection.ReplaceOne(ctx, filter, user)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete removes an external user
func (r *mongoExternalRepository) Delete(ctx context.Context, id string) error {
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

// List returns all external users
func (r *mongoExternalRepository) List(ctx context.Context) ([]*model.ExternalUser, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.ExternalUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// ListByDepartment returns all external users from a specific department
func (r *mongoExternalRepository) ListByDepartment(ctx context.Context, department string) ([]*model.ExternalUser, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"company.department": department}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.ExternalUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// ImportFromAPI imports external users from an API and saves them to database
func (r *mongoExternalRepository) ImportFromAPI(ctx context.Context, apiURL string) (int, error) {
	// Default to dummyjson if no URL provided
	if apiURL == "" {
		apiURL = "https://dummyjson.com/users"
	}

	// Fetch data from external API
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Parse response
	var response struct {
		Users []map[string]interface{} `json:"users"`
		Total int                      `json:"total"`
		Skip  int                      `json:"skip"`
		Limit int                      `json:"limit"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}

	// Convert users to our model and save to database
	collection := r.client.Database(r.database).Collection(r.collection)

	// Clear existing data (optional)
	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear existing data: %v", err)
	}

	// Prepare bulk insert
	var operations []mongo.WriteModel
	for _, userData := range response.Users {
		// Add timestamp
		userData["created_at"] = time.Now()

		// Create write model for bulk insert
		operation := mongo.NewInsertOneModel().SetDocument(userData)
		operations = append(operations, operation)
	}

	// Execute bulk write if there are operations
	if len(operations) > 0 {
		_, err = collection.BulkWrite(ctx, operations)
		if err != nil {
			return 0, err
		}
	}

	return len(operations), nil
}

// Disconnect is a no-op since we don't manage the connection
func (r *mongoExternalRepository) Disconnect(ctx context.Context) error {
	// We don't disconnect here since the client is managed externally
	return nil
}
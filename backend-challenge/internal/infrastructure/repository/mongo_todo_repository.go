package repository

import (
	"context"
	"errors"
	"time"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Domain errors
var (
	ErrTodoNotFound = errors.New("todo item not found")
)

// mongoTodoRepository implements the TodoRepository interface
type mongoTodoRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoTodoRepository creates a new MongoDB repository for todo items
func NewMongoTodoRepository(client *mongo.Client, dbName string) repository.TodoRepository {
	return &mongoTodoRepository{
		client:     client,
		database:   dbName,
		collection: "todos",
	}
}

// Create adds a new todo item
func (r *mongoTodoRepository) Create(ctx context.Context, todo *model.TodoItem) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Set timestamps if not already set
	now := time.Now()
	if todo.CreatedAt.IsZero() {
		todo.CreatedAt = now
	}
	if todo.UpdatedAt.IsZero() {
		todo.UpdatedAt = now
	}

	// Insert the document
	_, err := collection.InsertOne(ctx, todo)
	return err
}

// GetByID fetches a todo item by ID
func (r *mongoTodoRepository) GetByID(ctx context.Context, id string) (*model.TodoItem, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"_id": id}

	var todo model.TodoItem
	err := collection.FindOne(ctx, filter).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}

	return &todo, nil
}

// Update updates a todo item
func (r *mongoTodoRepository) Update(ctx context.Context, todo *model.TodoItem) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Update the document
	filter := bson.M{"_id": todo.ID}
	update := bson.M{
		"$set": bson.M{
			"type":       todo.Type,
			"name":       todo.Name,
			"status":     todo.Status,
			"clicked_at": todo.ClickedAt,
			"return_at":  todo.ReturnAt,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrTodoNotFound
	}

	return nil
}

// Delete removes a todo item
func (r *mongoTodoRepository) Delete(ctx context.Context, id string) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"_id": id}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrTodoNotFound
	}

	return nil
}

// List returns all todo items
func (r *mongoTodoRepository) List(ctx context.Context) ([]*model.TodoItem, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Empty filter to get all documents
	filter := bson.M{}

	// Set up find options to sort by creation date
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []*model.TodoItem
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

// FindByStatus returns all todo items with a specific status
func (r *mongoTodoRepository) FindByStatus(ctx context.Context, status model.ItemStatus) ([]*model.TodoItem, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"status": status}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []*model.TodoItem
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

// FindByTypeAndStatus returns all todo items with a specific type and status
func (r *mongoTodoRepository) FindByTypeAndStatus(ctx context.Context, itemType model.ItemType, status model.ItemStatus) ([]*model.TodoItem, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{
		"type":   itemType,
		"status": status,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []*model.TodoItem
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

// UpdateStatus updates the status of a todo item
func (r *mongoTodoRepository) UpdateStatus(ctx context.Context, id string, status model.ItemStatus) error {
	collection := r.client.Database(r.database).Collection(r.collection)

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrTodoNotFound
	}

	return nil
}

// FindToReturn finds all todo items that should be returned to the main list
func (r *mongoTodoRepository) FindToReturn(ctx context.Context, currentTime string) ([]*model.TodoItem, error) {
	collection := r.client.Database(r.database).Collection(r.collection)

	// Parse the current time
	now, err := time.Parse(time.RFC3339, currentTime)
	if err != nil {
		return nil, err
	}

	// Find items that should return to main list
	filter := bson.M{
		"status":    model.StatusColumn,
		"return_at": bson.M{"$lte": now},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []*model.TodoItem
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}
package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ModelPrice struct {
	Input, Output float64
}

var modelPrices = map[string]ModelPrice{
	"gpt-4o": {
		Input:  5,
		Output: 15,
	},
	"gpt-4o-mini": {
		Input:  0.15,
		Output: 0.6,
	},
}

var perTokens = 1000000

type LogEntry struct {
	ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
	BatchID      string    `bson:"batchId" json:"batchId"`
	Timestamp    time.Time `bson:"timestamp" json:"timestamp"`
	Prompt       string    `bson:"prompt" json:"prompt"`
	Response     string    `bson:"response" json:"response"`
	InputTokens  int       `bson:"inputTokens" json:"inputTokens"`
	OutputTokens int       `bson:"outputTokens" json:"outputTokens"`
}

type Batch struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	TotalTokens int                `bson:"totalTokens" json:"totalTokens"`
	TotalCost   float64            `bson:"totalCost" json:"totalCost"`
}

type Logger struct {
	client *mongo.Client
	db     *mongo.Database
	logger *log.Logger
}

func NewLogger() (*Logger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	db := client.Database("tellm")
	logger := log.New(os.Stdout, "", log.LstdFlags)

	return &Logger{
		client: client,
		db:     db,
		logger: logger,
	}, nil
}

func (l *Logger) CreateBatch(id, description string) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid id: %v", err)
	}

	batch := Batch{
		ID:          objectID,
		Description: description,
		TotalTokens: 0,
		TotalCost:   0,
		CreatedAt:   time.Now(),
	}

	result, err := l.db.Collection("batches").InsertOne(ctx, batch)
	if err != nil {
		return primitive.NilObjectID, err
	}

	l.logger.Printf("Created new batch: %s\n", id)
	return result.InsertedID.(primitive.ObjectID), nil
}

func (l *Logger) Log(batchID, prompt, response, model string, inputTokens, outputTokens int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	entry := LogEntry{
		BatchID:      batchID,
		Timestamp:    time.Now(),
		Prompt:       prompt,
		Response:     response,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}

	_, err := l.db.Collection("logs").InsertOne(ctx, entry)
	if err != nil {
		return err
	}

	pricing := modelPrices[model]
	if pricing == (ModelPrice{}) {
		return fmt.Errorf("model not found: %s", model)
	}
	totalTokens := inputTokens + outputTokens
	totalCost := calculateCost(pricing, inputTokens, outputTokens)

	objectID, err := primitive.ObjectIDFromHex(batchID)
	if err != nil {
		return fmt.Errorf("invalid batchID: %v", err)
	}

	_, err = l.db.Collection("batches").UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$inc": bson.M{
				"totalTokens": totalTokens,
				"totalCost":   totalCost,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update batch: %v", err)
	}

	l.logger.Printf("Logged entry to batch: %s\n", batchID)
	return nil
}

func (l *Logger) GetLogs(batchID string) ([]LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"batchId": batchID}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := l.db.Collection("logs").Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []LogEntry
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	l.logger.Printf("Retrieved %d logs for batch: %s", len(logs), batchID)
	return logs, nil
}

func (l *Logger) GetBatches() ([]Batch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var batches []Batch
	cursor, err := l.db.Collection("batches").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}

	l.logger.Printf("Retrieved %d batches\n", len(batches))
	return batches, nil
}

func (l *Logger) GetBatch(batchID string) (Batch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(batchID)
	if err != nil {
		return Batch{}, fmt.Errorf("invalid batchID: %v", err)
	}

	var batch Batch
	err = l.db.Collection("batches").FindOne(ctx, bson.M{"_id": objectID}).Decode(&batch)
	if err != nil {
		return Batch{}, err
	}

	l.logger.Printf("Retrieved batch: %s\n", batchID)
	return batch, nil
}

func (l *Logger) Close() {
	if l.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		l.client.Disconnect(ctx)
	}
}

// Implement this function based on your pricing model
func calculateCost(pricing ModelPrice, inputTokens, outputTokens int) float64 {
	totalCost := (float64(inputTokens) / float64(perTokens)) * pricing.Input
	totalCost += (float64(outputTokens) / float64(perTokens)) * pricing.Output
	return totalCost
}

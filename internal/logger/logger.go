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
	ID          string    `bson:"_id" json:"id"`
	Description string    `bson:"description" json:"description"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
	TotalTokens int       `bson:"totalTokens" json:"totalTokens"`
	TotalCost   float64   `bson:"totalCost" json:"totalCost"`
}

type Logger struct {
	client *mongo.Client
	db     *mongo.Database
	logger *log.Logger
}

func NewLogger() (*Logger, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
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
	batch := Batch{
		ID:          id,
		Description: description,
		CreatedAt:   time.Now(),
	}

	result, err := l.db.Collection("batches").InsertOne(context.TODO(), batch)
	if err != nil {
		return primitive.NilObjectID, err
	}

	l.logger.Printf("Created new batch: %s\n", id)
	return result.InsertedID.(primitive.ObjectID), nil
}

func (l *Logger) Log(batchID, prompt, response, model string, inputTokens, outputTokens int) error {
	entry := LogEntry{
		BatchID:      batchID,
		Timestamp:    time.Now(),
		Prompt:       prompt,
		Response:     response,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}

	_, err := l.db.Collection("logs").InsertOne(context.Background(), entry)
	if err != nil {
		return err
	}

	pricing := modelPrices[model]
	if pricing == (ModelPrice{}) {
		return fmt.Errorf("model not found: %s", model)
	}
	totalTokens := inputTokens + outputTokens
	totalCost := calculateCost(pricing, inputTokens, outputTokens)

	opts := options.Update().SetUpsert(true)
	_, err = l.db.Collection("batches").UpdateOne(
		context.Background(),
		bson.M{"batchId": batchID},
		bson.M{
			"$set": bson.M{
				"createdAt": time.Now(),
			},
			"$inc": bson.M{
				"totalTokens": totalTokens,
				"totalCost":   totalCost,
			},
		},
		opts,
	)
	if err != nil {
		return err
	}

	l.logger.Printf("Logged entry to batch: %s\n", batchID)
	return nil
}

func (l *Logger) GetLogs(batchID string) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(batchID)
	if err != nil {
		return nil, fmt.Errorf("invalid batchID: %v", err)
	}

	var logs []LogEntry
	cursor, err := l.db.Collection("logs").Find(context.TODO(), bson.M{"_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return nil, err
	}

	var batch Batch
	err = l.db.Collection("batches").FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&batch)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"logs": logs,
		"batchInfo": map[string]interface{}{
			"totalTokens": batch.TotalTokens,
			"totalCost":   batch.TotalCost,
		},
	}

	l.logger.Printf("Retrieved %d logs and batch info for batch: %s\n", len(logs), batchID)
	return result, nil
}

func (l *Logger) GetBatches() ([]Batch, error) {
	var batches []Batch
	cursor, err := l.db.Collection("batches").Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &batches); err != nil {
		return nil, err
	}

	l.logger.Printf("Retrieved %d batches\n", len(batches))
	return batches, nil
}

func (l *Logger) Close() {
	if l.client != nil {
		l.client.Disconnect(context.TODO())
	}
}

// Implement this function based on your pricing model
func calculateCost(pricing ModelPrice, inputTokens, outputTokens int) float64 {
	totalCost := (float64(inputTokens) / float64(perTokens)) * pricing.Input
	totalCost += (float64(outputTokens) / float64(perTokens)) * pricing.Output
	return totalCost
}

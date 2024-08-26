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

type ModelInfo struct {
	Name   string  `bson:"name" json:"name"`
	Lab    string  `bson:"lab" json:"lab"`
	Input  float64 `bson:"input" json:"input"`
	Output float64 `bson:"output" json:"output"`
}

type ModelInfoMap map[string]ModelInfo

var modelPrices ModelInfoMap

func init() {
	modelPrices = ModelInfoMap{
		"gpt-4o": {
			Name:   "gpt-4o",
			Lab:    "OpenAI",
			Input:  5,
			Output: 15,
		},
		"gpt-4o-mini": {
			Name:   "gpt-4o-mini",
			Lab:    "OpenAI",
			Input:  0.15,
			Output: 0.6,
		},
		"claude-3-5-sonnet-20240620": {
			Name:   "claude-3-5-sonnet-20240620",
			Lab:    "Anthropic",
			Input:  3,
			Output: 15,
		},
		"gpt-4o-2024-08-06": {
			Name:   "gpt-4o-2024-08-06",
			Lab:    "OpenAI",
			Input:  2.5,
			Output: 10,
		},
	}
}

func (m ModelInfoMap) GetModelInfo(model string) (ModelInfo, error) {
	info, exists := m[model]
	if !exists {
		return ModelInfo{}, fmt.Errorf("model not found: %s", model)
	}
	return info, nil
}

var perTokens = 1000000

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

func (l *Logger) CreateBatch(ctx context.Context, batchReq BatchRequest) (primitive.ObjectID, error) {
	objectID, err := primitive.ObjectIDFromHex(batchReq.ID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid id: %v", err)
	}

	batch := Batch{
		ID:          objectID,
		CreatedBy:   batchReq.CreatedBy,
		Name:        batchReq.Name,
		Plan:        "free",
		TotalTokens: 0,
		InputCost:   0,
		OutputCost:  0,
		CreatedAt:   time.Now(),
	}

	result, err := l.db.Collection("batches").InsertOne(ctx, batch)
	if err != nil {
		return primitive.NilObjectID, err
	}

	l.logger.Printf("Created new batch: %s\n", batchReq.ID)
	return result.InsertedID.(primitive.ObjectID), nil
}

func (l *Logger) Log(ctx context.Context, req EntryRequest) error {
	var objectID primitive.ObjectID
	_, err := l.GetBatch(ctx, req.BatchID)
	if err != nil {
		l.logger.Printf("Batch with id %s not found, creating new one\n", req.BatchID)
		objectID, err = l.CreateBatch(ctx, BatchRequest{
			ID:        req.BatchID,
			Name:      req.Name,
			CreatedBy: req.CreatedBy,
		})
		if err != nil {
			return err
		}
	}

	modelInfo, err := modelPrices.GetModelInfo(req.Model)
	if err != nil {
		return err
	}

	inputCost := calculateCost(modelInfo.Input, req.InputTokens)
	outputCost := calculateCost(modelInfo.Output, req.OutputTokens)

	entry := LogEntry{
		BatchID:      req.BatchID,
		Timestamp:    time.Now(),
		CreatedBy:    req.CreatedBy,
		Prompt:       req.Prompt,
		Name:         req.Name,
		Response:     req.Response,
		ModelInfo:    modelInfo,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		InputTokens:  req.InputTokens,
		OutputTokens: req.OutputTokens,
	}

	_, err = l.db.Collection("logs").InsertOne(ctx, entry)
	if err != nil {
		return err
	}

	totalTokens := req.InputTokens + req.OutputTokens

	if objectID == primitive.NilObjectID {
		objectID, err = primitive.ObjectIDFromHex(req.BatchID)
		if err != nil {
			return fmt.Errorf("invalid batchID: %v", err)
		}
	}

	updateFields := bson.M{
		"$inc": bson.M{
			"totalTokens": totalTokens,
			"inputCost":   inputCost,
			"outputCost":  outputCost,
		},
	}

	if modelInfo.Name == "gpt-4o-2024-08-06" {
		updateFields["$set"] = bson.M{"plan": "premium"}
	}

	_, err = l.db.Collection("batches").UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		updateFields,
	)
	if err != nil {
		return fmt.Errorf("failed to update batch: %v", err)
	}

	l.logger.Printf("Logged entry to batch: %s\n", req.BatchID)
	return nil
}

func (l *Logger) GetLogs(ctx context.Context, batchID string) ([]LogEntry, error) {
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

func (l *Logger) GetBatches(ctx context.Context) ([]Batch, error) {
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
	if batches == nil {
		return []Batch{}, nil
	}
	return batches, nil
}

func (l *Logger) GetBatch(ctx context.Context, batchID string) (Batch, error) {
	objectID, err := primitive.ObjectIDFromHex(batchID)
	if err != nil {
		return Batch{}, fmt.Errorf("invalid batchID: %w", err)
	}

	var batch Batch
	err = l.db.Collection("batches").FindOne(ctx, bson.M{"_id": objectID}).Decode(&batch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Batch{}, fmt.Errorf("batch not found: %s", batchID)
		}
		return Batch{}, fmt.Errorf("error retrieving batch: %w", err)
	}

	l.logger.Printf("Retrieved batch: %s", batchID)
	return batch, nil
}

func (l *Logger) Close(ctx context.Context) {
	if l.client != nil {
		l.client.Disconnect(ctx)
	}
}

func calculateCost(pricePerMillion float64, tokens int) float64 {
	return (float64(tokens) / float64(perTokens)) * pricePerMillion
}

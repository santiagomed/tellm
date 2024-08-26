package logger

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntryRequest struct {
	BatchID      string `json:"batch_id"`
	CreatedBy    string `json:"created_by"`
	Name         string `json:"name"`
	Prompt       string `json:"prompt"`
	Response     string `json:"response"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

type BatchRequest struct {
	ID        string `json:"id"`
	CreatedBy string `json:"created_by"`
	Name      string `json:"name"`
}

type LogEntry struct {
	ID           string    `bson:"_id,omitempty" json:"id,omitempty"`
	BatchID      string    `bson:"batchId" json:"batchId"`
	CreatedBy    string    `bson:"createdBy" json:"createdBy"`
	Timestamp    time.Time `bson:"timestamp" json:"timestamp"`
	Name         string    `bson:"name" json:"name"`
	Prompt       string    `bson:"prompt" json:"prompt"`
	Response     string    `bson:"response" json:"response"`
	ModelInfo    ModelInfo `bson:"modelInfo" json:"modelInfo"`
	InputTokens  int       `bson:"inputTokens" json:"inputTokens"`
	InputCost    float64   `bson:"inputCost" json:"inputCost"`
	OutputTokens int       `bson:"outputTokens" json:"outputTokens"`
	OutputCost   float64   `bson:"outputCost" json:"outputCost"`
}

type Batch struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	CreatedBy   string             `bson:"createdBy" json:"createdBy"`
	Name        string             `bson:"name" json:"name"`
	Plan        string             `bson:"plan" json:"plan"`
	TotalTokens int                `bson:"totalTokens" json:"totalTokens"`
	InputCost   float64            `bson:"inputCost" json:"inputCost"`
	OutputCost  float64            `bson:"outputCost" json:"outputCost"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

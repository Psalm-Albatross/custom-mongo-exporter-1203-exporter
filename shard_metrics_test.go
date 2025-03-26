package main

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCollectShardMetrics(t *testing.T) {
	uri := "mongodb://localhost:27017"

	// Mock MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Call the function
	collectShardMetrics(client, ctx)

	// Validate metrics
	if mongoShardCount == nil {
		t.Error("mongoShardCount metric is not initialized")
	}
	if mongoShardStats == nil {
		t.Error("mongoShardStats metric is not initialized")
	}
}

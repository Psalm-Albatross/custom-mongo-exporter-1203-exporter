package main

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCollectCollectionMetrics(t *testing.T) {
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
	collectCollectionMetrics(client, ctx)

	// Validate metrics
	if mongoTotalCollections == nil {
		t.Error("mongoTotalCollections metric is not initialized")
	}
	if mongoTotalDocuments == nil {
		t.Error("mongoTotalDocuments metric is not initialized")
	}
}

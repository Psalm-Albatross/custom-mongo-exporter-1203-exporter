package main

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCollectMongoMetrics(t *testing.T) {
	uri := "mongodb://localhost:27017"
	user := ""
	password := ""

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
	collectMongoMetrics(uri, user, password)

	// Validate metrics
	if mongoUp == nil {
		t.Error("mongoUp metric is not initialized")
	}
	if mongoDBCount == nil {
		t.Error("mongoDBCount metric is not initialized")
	}
}

func TestCalculateTPS(t *testing.T) {
	currentOpCounters := map[string]int64{
		"insert":  100,
		"update":  200,
		"delete":  300,
		"query":   400,
		"command": 500,
	}
	intervalSeconds := 10.0

	// Call the function
	calculateTPS(currentOpCounters, intervalSeconds)

	// Validate TPS metrics
	for opType, value := range currentOpCounters {
		if value <= 0 {
			t.Errorf("TPS for %s is not calculated correctly", opType)
		}
	}
}

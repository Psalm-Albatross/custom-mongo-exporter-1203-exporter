package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Mock MongoDB client for testing
type mockMongoClient struct {
	pingError       error
	serverStatus    bson.M
	serverStatusErr error
}

func (m *mockMongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error { // Use readpref.ReadPref
	return m.pingError
}

func (m *mockMongoClient) Database(name string, opts ...*options.DatabaseOptions) *mockDatabase {
	return &mockDatabase{
		serverStatus:    m.serverStatus,
		serverStatusErr: m.serverStatusErr,
	}
}

type mockDatabase struct {
	serverStatus    bson.M
	serverStatusErr error
}

func (db *mockDatabase) RunCommand(ctx context.Context, cmd interface{}, opts ...*options.RunCmdOptions) *mockSingleResult {
	return &mockSingleResult{
		result: db.serverStatus,
		err:    db.serverStatusErr,
	}
}

type mockSingleResult struct {
	result bson.M
	err    error
}

func (sr *mockSingleResult) Decode(v interface{}) error {
	if sr.err != nil {
		return sr.err
	}
	*v.(*bson.M) = sr.result
	return nil
}

var testCollectMongoMetrics func(uri, user, password string)

func TestCollectMongoMetrics_Success(t *testing.T) {
	mockClient := &mockMongoClient{
		pingError: nil,
		serverStatus: bson.M{
			"opcounters": bson.M{"insert": int32(10), "query": int32(20)},
			"mem":        bson.M{"resident": int32(100), "page_faults": int32(5)},
		},
	}

	testCollectMongoMetrics = func(uri, user, password string) {
		client := mockClient
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		serverStatus := bson.M{}
		if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Validate metrics
		if testutil.ToFloat64(mongoUp) != 1 {
			t.Errorf("Expected mongoUp to be 1, got %v", testutil.ToFloat64(mongoUp))
		}
		if testutil.ToFloat64(mongoOpCounters.WithLabelValues("insert")) != 10 {
			t.Errorf("Expected insert counter to be 10, got %v", testutil.ToFloat64(mongoOpCounters.WithLabelValues("insert")))
		}
		if testutil.ToFloat64(mongoMemoryUsage) != 100*1024*1024 {
			t.Errorf("Expected memory usage to be 100MB, got %v", testutil.ToFloat64(mongoMemoryUsage))
		}
	}
}

func TestCollectMongoMetrics_Failure(t *testing.T) {
	mockClient := &mockMongoClient{
		pingError: errors.New("failed to connect"),
	}

	testCollectMongoMetrics = func(uri, user, password string) {
		client := mockClient
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx, readpref.Primary()); err == nil {
			t.Errorf("Expected error, got nil")
		}

		// Validate metrics
		if testutil.ToFloat64(mongoUp) != 0 {
			t.Errorf("Expected mongoUp to be 0, got %v", testutil.ToFloat64(mongoUp))
		}
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Define Prometheus metrics
	mongoUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mongodb_up",
		Help: "MongoDB is up and reachable",
	})

	mongoTotalCollections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mongodb_total_collections",
		Help: "Total number of collections in MongoDB",
	})

	mongoTotalDataSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "mongodb_total_data_size",
		Help: "Total data size in MongoDB in bytes",
	})
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(mongoUp)
	prometheus.MustRegister(mongoTotalCollections)
	prometheus.MustRegister(mongoTotalDataSize)
}

// collectMetrics connects to MongoDB and updates the metrics
func collectMetrics(uri string) {
	clientOpts := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		mongoUp.Set(0)
		return
	}

	// Ping MongoDB to check if it's reachable
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		mongoUp.Set(0)
		return
	}

	mongoUp.Set(1)

	// Get database stats
	stats, err := client.Database("admin").RunCommand(context.TODO(), mongo.M{"dbStats": 1}).DecodeBytes()
	if err != nil {
		log.Printf("Error getting database stats: %v", err)
		return
	}

	// Update the total collections and data size metrics
	collections, _ := stats.Lookup("collections").AsInt64OK()
	dataSize, _ := stats.Lookup("dataSize").AsInt64OK()

	mongoTotalCollections.Set(float64(collections))
	mongoTotalDataSize.Set(float64(dataSize))

	// Close the MongoDB connection
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}
}

func main() {
	// MongoDB URI, replace with your MongoDB instance URI
	mongoURI := "mongodb://localhost:27017"

	// Create a ticker to collect metrics periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Collect metrics in a separate goroutine
	go func() {
		for range ticker.C {
			collectMetrics(mongoURI)
		}
	}()

	// Expose the Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start the HTTP server for Prometheus to scrape metrics
	log.Println("Starting MongoDB Exporter on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

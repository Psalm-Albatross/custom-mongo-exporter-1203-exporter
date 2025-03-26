package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	// MongoDB Metrics
	mongoUp              = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_up", Help: "MongoDB is up and reachable"})
	mongoDBCount         = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_db_count", Help: "Total number of databases in MongoDB"})
	mongoDBList          = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_databases", Help: "List of databases in MongoDB"}, []string{"database"})
	mongoIdleConnections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_idle_connections", Help: "Number of idle connections to MongoDB"})
	mongoOpCounters      = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_op_counters", Help: "MongoDB operation counters by type"}, []string{"type"})
	mongoReplicationLag  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_replication_lag", Help: "Replication lag for each MongoDB node"}, []string{"replica_set_member"})
	mongoMemoryUsage     = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_memory_usage_bytes", Help: "MongoDB memory usage in bytes"})
	mongoCacheUsage      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_cache_usage_bytes", Help: "MongoDB cache usage in bytes"})
	mongoPageFaults      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_page_faults", Help: "MongoDB page faults"})

	// MongoDB Lock Metrics
	mongoLockTotalTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_lock_time_ms", Help: "Total lock time for read and write operations"}, []string{"lock_type"})
	mongoLockRatio     = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_lock_ratio", Help: "Lock ratio for read and write operations"}, []string{"lock_type"})

	// System Metrics
	cpuUsage    = prometheus.NewGauge(prometheus.GaugeOpts{Name: "system_cpu_usage_percent", Help: "CPU usage percentage"})
	memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{Name: "system_memory_usage_bytes", Help: "Memory usage in bytes"})
	diskUsage   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "system_disk_usage_bytes", Help: "Disk usage in bytes"})

	// Networking Metrics
	mongoNetworkIn          = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_network_in_bytes", Help: "Total bytes received by MongoDB"})
	mongoNetworkOut         = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_network_out_bytes", Help: "Total bytes sent by MongoDB"})
	mongoCurrentConnections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_current_connections", Help: "Current number of connections to MongoDB"})

	// Additional Metrics
	mongoTransactionsPerSecond = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_transactions_per_second", Help: "MongoDB transactions per second"})

	// Additional Metrics for DBA Insights
	mongoQueryExecutionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_query_execution_time_ms", Help: "MongoDB query execution time in milliseconds"}, []string{"collection"})
	mongoQueryLatency       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_query_latency_ms", Help: "MongoDB query latency in milliseconds"}, []string{"collection"})
	mongoQueryPerformance   = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_query_performance", Help: "MongoDB query performance metrics"}, []string{"collection", "metric"})

	// Additional Metrics for DBA Insights
	mongoIndexUsage        = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_index_usage", Help: "MongoDB index usage"}, []string{"index"})
	mongoDocumentInserts   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_document_inserts", Help: "Number of documents inserted"})
	mongoDocumentUpdates   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_document_updates", Help: "Number of documents updated"})
	mongoDocumentDeletes   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_document_deletes", Help: "Number of documents deleted"})
	mongoActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_active_connections", Help: "Number of active connections to MongoDB"})
	mongoTotalConnections  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_total_connections", Help: "Total number of connections to MongoDB"})
	mongoOpLatency         = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_op_latency_ms", Help: "MongoDB operation latency in milliseconds"}, []string{"operation"})

	// Additional Metrics for Disk I/O
	mongoDiskReadBytes  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_disk_read_bytes", Help: "Total bytes read from disk by MongoDB"})
	mongoDiskWriteBytes = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_disk_write_bytes", Help: "Total bytes written to disk by MongoDB"})

	// Additional Metrics for Shard Statistics
	mongoShardCount = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_shard_count", Help: "Number of shards in the MongoDB cluster"})
	mongoShardStats = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_shard_stats", Help: "Statistics for each shard"}, []string{"shard", "metric"})

	// Additional Metrics for Replica Set Health
	mongoReplicaSetHealth = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_replica_set_health", Help: "Health of each replica set member"}, []string{"member"})
	mongoReplicaSetState  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_replica_set_state", Help: "State of each replica set member"}, []string{"member"})

	// Additional Metrics for Detailed Operation Statistics
	mongoOperationStats = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_operation_stats", Help: "Detailed operation statistics"}, []string{"operation", "metric"})

	// Advanced Metrics
	mongoSlowQueries         = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_slow_queries", Help: "Number of slow queries"}, []string{"collection"})
	mongoConnectionPoolStats = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_connection_pool_stats", Help: "Connection pool statistics"}, []string{"pool", "metric"})
	mongoStorageEngineStats  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_storage_engine_stats", Help: "Storage engine statistics"}, []string{"engine", "metric"})

	// Additional Metrics for Collections and Documents
	mongoTotalCollections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_total_collections", Help: "Total number of collections across all databases"})
	mongoTotalDocuments   = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_total_documents", Help: "Total number of documents across all collections"})
	mongoCollectionStats  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_collection_stats", Help: "Statistics for each collection"}, []string{"database", "collection", "metric"})

	// Additional Metrics for Cache and Index Insights
	mongoCacheEvictions      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_cache_evictions", Help: "Number of cache evictions in MongoDB"})
	mongoIndexAccessPatterns = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mongodb_index_access_patterns", Help: "Access patterns for indexes"}, []string{"index"})

	// Additional Metrics for Client Connections
	mongoActiveClientConnections = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_active_client_connections", Help: "Number of active client connections to MongoDB"})
	mongoTotalClientConnections  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mongodb_total_client_connections", Help: "Total number of client connections to MongoDB"})

	// Flags for Optional Metrics
	disableDBListing      bool
	disableOpCounters     bool
	disableReplication    bool
	disableMemoryUsage    bool
	disableLockMetrics    bool
	disableNetworkMetrics bool
	// Version variable to check build version of custom-mongo-exporter tools
	version string

	// Customizable Collection Interval
	collectionInterval time.Duration

	// Metrics for tracking errors
	mongoErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "mongodb_error_count", Help: "Total number of MongoDB errors encountered"}, []string{"error_type"})
)

func init() {
	// Register all metrics
	prometheus.MustRegister(
		mongoUp, mongoDBCount, mongoDBList, mongoIdleConnections, mongoOpCounters,
		mongoReplicationLag, mongoMemoryUsage, mongoCacheUsage, mongoPageFaults,
		mongoLockTotalTime, mongoLockRatio, cpuUsage, memoryUsage, diskUsage,
		mongoNetworkIn, mongoNetworkOut, mongoCurrentConnections, mongoTransactionsPerSecond,
		mongoQueryExecutionTime, mongoQueryLatency, mongoQueryPerformance,
		mongoIndexUsage, mongoDocumentInserts, mongoDocumentUpdates, mongoDocumentDeletes,
		mongoActiveConnections, mongoTotalConnections, mongoOpLatency,
		mongoDiskReadBytes, mongoDiskWriteBytes, mongoErrorCount,
		mongoShardCount, mongoShardStats,
		mongoReplicaSetHealth, mongoReplicaSetState,
		mongoOperationStats,
		mongoSlowQueries,
		mongoConnectionPoolStats,
		mongoStorageEngineStats,
		mongoTotalCollections,
		mongoTotalDocuments,
		mongoCollectionStats,
		mongoCacheEvictions,
		mongoIndexAccessPatterns,
		mongoActiveClientConnections,
		mongoTotalClientConnections,
	)
}

// Variable to store previous opcounters for TPS calculation
var previousOpCounters = map[string]int64{
	"insert":  0,
	"update":  0,
	"delete":  0,
	"query":   0,
	"command": 0,
}

func calculateTPS(currentOpCounters map[string]int64, intervalSeconds float64) {
	log.Println("Calculating Transactions Per Second (TPS)...")
	for opType, currentValue := range currentOpCounters {
		if previousValue, exists := previousOpCounters[opType]; exists {
			tps := float64(currentValue-previousValue) / intervalSeconds
			mongoOpCounters.WithLabelValues(opType).Set(tps)
			log.Printf("TPS for %s: %f", opType, tps)
		} else {
			log.Printf("No previous value found for operation type: %s", opType)
		}
		previousOpCounters[opType] = currentValue
	}
	log.Println("TPS calculation completed.")
}

func collectMongoMetrics(uri, user, password string) {
	log.Println("Starting MongoDB metrics collection...")
	// Context with timeout to prevent hanging operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	if user != "" && password != "" {
		clientOpts.SetAuth(options.Credential{Username: user, Password: password})
		log.Println("Using MongoDB authentication with provided credentials.")
	} else {
		log.Println("No MongoDB credentials provided, attempting connection without authentication.")
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		mongoUp.Set(0)
		mongoErrorCount.WithLabelValues("connection").Inc()
		log.Println("MongoDB connection failed. Metrics collection aborted.")
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
			mongoErrorCount.WithLabelValues("disconnection").Inc()
		} else {
			log.Println("Disconnected from MongoDB successfully.")
		}
	}()

	// Check the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Printf("MongoDB ping failed: %v", err)
		mongoUp.Set(0)
		mongoErrorCount.WithLabelValues("ping").Inc()
		return
	}
	log.Println("Connected to MongoDB successfully!")
	mongoUp.Set(1)

	serverStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
		log.Printf("Failed to get server status: %v", err)
		mongoErrorCount.WithLabelValues("server_status").Inc()
		return
	}

	// Op Counters
	if !disableOpCounters {
		if opCounters, ok := serverStatus["opcounters"].(bson.M); ok {
			for opType, count := range opCounters {
				if cnt, ok := count.(int64); ok { // Fix: Ensure type is int64 for large values
					mongoOpCounters.WithLabelValues(opType).Set(float64(cnt))
				} else {
					log.Printf("Failed to parse op counter for type %s", opType)
				}
			}
			log.Println("MongoDB operation counters collected.")
		}
	}

	// Memory Usage
	if !disableMemoryUsage {
		if memInfo, ok := serverStatus["mem"].(bson.M); ok {
			if resident, ok := memInfo["resident"].(int64); ok { // Fix: Ensure type is int64
				mongoMemoryUsage.Set(float64(resident) * 1024 * 1024) // Convert MB to bytes
				log.Printf("MongoDB resident memory: %d MB", resident)
			} else {
				log.Println("Failed to parse resident memory.")
			}
			if faults, ok := memInfo["page_faults"].(int64); ok { // Fix: Ensure type is int64
				mongoPageFaults.Set(float64(faults))
				log.Printf("MongoDB page faults: %d", faults)
			} else {
				log.Println("Failed to parse page faults.")
			}
		}
	}

	// Lock Metrics
	if !disableLockMetrics {
		if locks, ok := serverStatus["locks"].(bson.M); ok {
			for lockType, lockInfo := range locks {
				if lockData, ok := lockInfo.(bson.M); ok {
					if timeLocked, ok := lockData["timeLockedMicros"].(bson.M); ok {
						for opType, value := range timeLocked {
							if val, ok := value.(int64); ok {
								mongoLockTotalTime.WithLabelValues(lockType + "_" + opType).Set(float64(val) / 1000) // microseconds to milliseconds
							}
						}
					}
				}
			}
			log.Println("MongoDB lock metrics collected.")
		}
	}

	// Transactions Per Second
	if tps, ok := serverStatus["opcounters"].(bson.M); ok {
		currentOpCounters := map[string]int64{}
		for opType, value := range tps {
			if count, ok := value.(int64); ok {
				currentOpCounters[opType] = count
			}
		}
		// Calculate TPS using the current and previous opcounters
		calculateTPS(currentOpCounters, collectionInterval.Seconds())
	}

	// Query Execution Time and Latency
	if metrics, ok := serverStatus["metrics"].(bson.M); ok {
		if commands, ok := metrics["commands"].(bson.M); ok {
			for cmd, cmdInfo := range commands {
				if cmdData, ok := cmdInfo.(bson.M); ok {
					if executionTime, ok := cmdData["totalMillis"].(int64); ok { // Fix: Ensure type is int64
						mongoQueryExecutionTime.WithLabelValues(cmd).Set(float64(executionTime))
						log.Printf("MongoDB query execution time for %s: %d ms", cmd, executionTime)
					} else {
						log.Printf("Execution time not found for command: %s", cmd)
					}
					if latency, ok := cmdData["latencyMillis"].(int64); ok { // Fix: Ensure type is int64
						mongoQueryLatency.WithLabelValues(cmd).Set(float64(latency))
						log.Printf("MongoDB query latency for %s: %d ms", cmd, latency)
					} else {
						log.Printf("Latency not found for command: %s", cmd)
					}
				}
			}
		}
	}

	// Query Performance Metrics
	if metrics, ok := serverStatus["metrics"].(bson.M); ok {
		if queryExecutor, ok := metrics["queryExecutor"].(bson.M); ok {
			for metric, value := range queryExecutor {
				if val, ok := value.(int32); ok {
					mongoQueryPerformance.WithLabelValues("queryExecutor", metric).Set(float64(val))
					log.Printf("MongoDB query performance metric %s: %d", metric, val)
				}
			}
		}
	}

	// Index Usage
	if indexCounters, ok := serverStatus["indexCounters"].(bson.M); ok {
		if accesses, ok := indexCounters["accesses"].(bson.M); ok {
			for index, count := range accesses {
				if cnt, ok := count.(int32); ok {
					mongoIndexUsage.WithLabelValues(index).Set(float64(cnt))
					log.Printf("MongoDB index usage for %s: %d", index, cnt)
				}
			}
		}
	}

	// Document Metrics
	if metrics, ok := serverStatus["metrics"].(bson.M); ok {
		if document, ok := metrics["document"].(bson.M); ok {
			if inserts, ok := document["inserted"].(int64); ok { // Fix: Ensure type is int64
				mongoDocumentInserts.Set(float64(inserts))
				log.Printf("MongoDB document inserts: %d", inserts)
			} else {
				log.Println("MongoDB document inserts metric not found.")
			}
			if updates, ok := document["updated"].(int64); ok { // Fix: Ensure type is int64
				mongoDocumentUpdates.Set(float64(updates))
				log.Printf("MongoDB document updates: %d", updates)
			} else {
				log.Println("MongoDB document updates metric not found.")
			}
			if deletes, ok := document["deleted"].(int64); ok { // Fix: Ensure type is int64
				mongoDocumentDeletes.Set(float64(deletes))
				log.Printf("MongoDB document deletes: %d", deletes)
			} else {
				log.Println("MongoDB document deletes metric not found.")
			}
		} else {
			log.Println("MongoDB document metrics not found in serverStatus.")
		}
	} else {
		log.Println("MongoDB metrics section not found in serverStatus.")
	}

	// Connection Metrics
	if connections, ok := serverStatus["connections"].(bson.M); ok {
		if active, ok := connections["current"].(int64); ok { // Fix: Ensure type is int64
			mongoActiveConnections.Set(float64(active))
			log.Printf("MongoDB active connections: %d", active)
		} else {
			log.Println("Failed to parse active connections.")
		}
		if total, ok := connections["totalCreated"].(int64); ok { // Fix: Ensure type is int64
			mongoTotalConnections.Set(float64(total))
			log.Printf("MongoDB total connections: %d", total)
		} else {
			log.Println("Failed to parse total connections.")
		}
	}

	// Operation Latency
	if opLatencies, ok := serverStatus["opLatencies"].(bson.M); ok {
		if reads, ok := opLatencies["reads"].(bson.M); ok {
			if latency, ok := reads["latency"].(int64); ok { // Fix: Ensure type is int64
				mongoOpLatency.WithLabelValues("read").Set(float64(latency))
				log.Printf("MongoDB read operation latency: %d ms", latency)
			}
		}
		if writes, ok := opLatencies["writes"].(bson.M); ok {
			if latency, ok := writes["latency"].(int64); ok { // Fix: Ensure type is int64
				mongoOpLatency.WithLabelValues("write").Set(float64(latency))
				log.Printf("MongoDB write operation latency: %d ms", latency)
			}
		}
		if commands, ok := opLatencies["commands"].(bson.M); ok {
			if latency, ok := commands["latency"].(int64); ok { // Fix: Ensure type is int64
				mongoOpLatency.WithLabelValues("command").Set(float64(latency))
				log.Printf("MongoDB command operation latency: %d ms", latency)
			}
		}
	}

	// Disk I/O Metrics
	if metrics, ok := serverStatus["metrics"].(bson.M); ok {
		if disk, ok := metrics["disk"].(bson.M); ok {
			if readBytes, ok := disk["readBytes"].(int64); ok { // Fix: Ensure type is int64
				mongoDiskReadBytes.Set(float64(readBytes))
				log.Printf("MongoDB disk read bytes: %d", readBytes)
			} else {
				log.Println("Failed to parse disk read bytes.")
			}
			if writeBytes, ok := disk["writeBytes"].(int64); ok { // Fix: Ensure type is int64
				mongoDiskWriteBytes.Set(float64(writeBytes))
				log.Printf("MongoDB disk write bytes: %d", writeBytes)
			} else {
				log.Println("Failed to parse disk write bytes.")
			}
		}
	}

	// DB Count and Listing
	if !disableDBListing {
		dbs, err := client.ListDatabaseNames(ctx, bson.M{})
		if err != nil {
			log.Printf("Failed to list databases: %v", err)
			mongoErrorCount.WithLabelValues("list_databases").Inc()
		} else {
			mongoDBCount.Set(float64(len(dbs)))
			for _, db := range dbs {
				mongoDBList.WithLabelValues(db).Set(1)
			}
			log.Println("Database listing collected.")
		}
	}

	// Replication Lag
	if !disableReplication {
		replStatus := bson.M{}
		if err := client.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1}).Decode(&replStatus); err != nil {
			log.Printf("Failed to get replication status: %v", err)
			mongoErrorCount.WithLabelValues("replication_status").Inc()
		} else {
			if members, ok := replStatus["members"].(primitive.A); ok {
				for _, member := range members {
					if memberInfo, ok := member.(bson.M); ok {
						if lag, ok := memberInfo["optimeDate"].(primitive.DateTime); ok {
							mongoReplicationLag.WithLabelValues(memberInfo["name"].(string)).Set(float64(time.Since(lag.Time()).Seconds()))
						}
					}
				}
				log.Println("Replication lag collected.")
			}
		}
	}

	// Idle Connections
	if !disableNetworkMetrics {
		serverStatus := bson.M{}
		if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
			log.Printf("Failed to collect idle connections: %v", err)
			mongoErrorCount.WithLabelValues("idle_connections").Inc()
		} else {
			if connections, ok := serverStatus["connections"].(bson.M); ok {
				if idle, ok := connections["available"].(int32); ok {
					mongoIdleConnections.Set(float64(idle))
					log.Printf("Idle connections: %d", idle)
				}
			}
		}
	}

	// Collect Networking Metrics
	if !disableNetworkMetrics {
		networkStats, err := net.IOCounters(true)
		if err != nil {
			log.Printf("Failed to get network stats: %v", err)
			mongoErrorCount.WithLabelValues("network_stats").Inc()
			return
		}
		for _, netStat := range networkStats {
			mongoNetworkIn.Set(float64(netStat.BytesRecv))
			mongoNetworkOut.Set(float64(netStat.BytesSent))
			mongoCurrentConnections.Set(float64(netStat.PacketsSent + netStat.PacketsRecv)) // A simple estimate of current connections
			log.Printf("MongoDB network in: %d bytes, network out: %d bytes", netStat.BytesRecv, netStat.BytesSent)
		}
	}
	log.Println("MongoDB metrics collection completed.")
}

func collectShardMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting shard metrics...")
	shardStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"listShards": 1}).Decode(&shardStatus); err != nil {
		log.Printf("Failed to get shard status: %v", err)
		mongoErrorCount.WithLabelValues("shard_status").Inc()
		return
	}

	if shards, ok := shardStatus["shards"].(primitive.A); ok {
		mongoShardCount.Set(float64(len(shards)))
		for _, shard := range shards {
			if shardInfo, ok := shard.(bson.M); ok {
				shardName := shardInfo["_id"].(string)
				// Example: Add shard-specific metrics
				mongoShardStats.WithLabelValues(shardName, "example_metric").Set(1) // Replace with actual metrics
			}
		}
		log.Println("Shard metrics collected.")
	}
	log.Println("Shard metrics collection completed.")
}

func collectReplicaSetMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting replica set metrics...")
	replStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1}).Decode(&replStatus); err != nil {
		log.Printf("Failed to get replica set status: %v", err)
		mongoErrorCount.WithLabelValues("replica_set_status").Inc()
		return
	}

	if members, ok := replStatus["members"].(primitive.A); ok {
		for _, member := range members {
			if memberInfo, ok := member.(bson.M); ok {
				memberName := memberInfo["name"].(string)
				if health, ok := memberInfo["health"].(int32); ok {
					mongoReplicaSetHealth.WithLabelValues(memberName).Set(float64(health))
				}
				if state, ok := memberInfo["state"].(int32); ok {
					mongoReplicaSetState.WithLabelValues(memberName).Set(float64(state))
				}
			}
		}
		log.Println("Replica set metrics collected.")
	}
	log.Println("Replica set metrics collection completed.")
}

func collectDetailedOperationMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting detailed operation metrics...")
	serverStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
		log.Printf("Failed to get server status for operation metrics: %v", err)
		mongoErrorCount.WithLabelValues("operation_metrics").Inc()
		return
	}

	if metrics, ok := serverStatus["metrics"].(bson.M); ok {
		if operation, ok := metrics["operation"].(bson.M); ok {
			for opType, opData := range operation {
				if opStats, ok := opData.(bson.M); ok {
					for metric, value := range opStats {
						if val, ok := value.(int32); ok {
							mongoOperationStats.WithLabelValues(opType, metric).Set(float64(val))
						}
					}
				}
			}
		}
		log.Println("Detailed operation metrics collected.")
	}
	log.Println("Detailed operation metrics collection completed.")
}

func collectAdvancedMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting advanced metrics...")
	// Collect Slow Queries
	currentOp := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"currentOp": 1, "active": true}).Decode(&currentOp); err != nil {
		log.Printf("Failed to get current operations: %v", err)
		mongoErrorCount.WithLabelValues("current_op").Inc()
	} else {
		if inprog, ok := currentOp["inprog"].(primitive.A); ok {
			for _, op := range inprog {
				if opInfo, ok := op.(bson.M); ok {
					if ns, ok := opInfo["ns"].(string); ok && ns != "" {
						if millis, ok := opInfo["millis"].(int32); ok && millis > 100 { // Assuming 100ms as the threshold for slow queries
							mongoSlowQueries.WithLabelValues(ns).Inc()
							log.Printf("Slow query detected on collection %s: %d ms", ns, millis)
						}
					}
				}
			}
		}
	}

	// Collect Connection Pool Statistics
	connPoolStats := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"connPoolStats": 1}).Decode(&connPoolStats); err != nil {
		log.Printf("Failed to get connection pool stats: %v", err)
		mongoErrorCount.WithLabelValues("conn_pool_stats").Inc()
	} else {
		for metric, value := range connPoolStats {
			if val, ok := value.(int32); ok {
				mongoConnectionPoolStats.WithLabelValues("default", metric).Set(float64(val))
				log.Printf("Connection pool metric %s: %d", metric, val)
			}
		}
	}

	// Collect Storage Engine Statistics
	serverStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
		log.Printf("Failed to get server status for storage engine stats: %v", err)
		mongoErrorCount.WithLabelValues("storage_engine_stats").Inc()
	} else {
		if storageEngine, ok := serverStatus["storageEngine"].(bson.M); ok {
			for metric, value := range storageEngine {
				if val, ok := value.(int32); ok {
					mongoStorageEngineStats.WithLabelValues("wiredTiger", metric).Set(float64(val))
					log.Printf("Storage engine metric %s: %d", metric, val)
				}
			}
		}
	}
	log.Println("Advanced metrics collection completed.")
}

func collectCollectionMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting collection metrics...")
	totalCollections := 0
	totalDocuments := int64(0)

	dbs, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to list databases: %v", err)
		mongoErrorCount.WithLabelValues("list_databases").Inc()
		return
	}

	for _, dbName := range dbs {
		collections, err := client.Database(dbName).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Printf("Failed to list collections for database %s: %v", dbName, err)
			mongoErrorCount.WithLabelValues("list_collections").Inc()
			continue
		}

		totalCollections += len(collections)

		for _, collectionName := range collections {
			collection := client.Database(dbName).Collection(collectionName)
			stats := bson.M{}
			if err := collection.Database().RunCommand(ctx, bson.M{"collStats": collectionName}).Decode(&stats); err != nil {
				log.Printf("Failed to get stats for collection %s.%s: %v", dbName, collectionName, err)
				mongoErrorCount.WithLabelValues("collection_stats").Inc()
				continue
			}

			if count, ok := stats["count"].(int64); ok {
				totalDocuments += count
				mongoCollectionStats.WithLabelValues(dbName, collectionName, "document_count").Set(float64(count))
			}

			if size, ok := stats["size"].(int64); ok {
				mongoCollectionStats.WithLabelValues(dbName, collectionName, "size_bytes").Set(float64(size))
			}

			if storageSize, ok := stats["storageSize"].(int64); ok {
				mongoCollectionStats.WithLabelValues(dbName, collectionName, "storage_size_bytes").Set(float64(storageSize))
			}

			if indexSize, ok := stats["totalIndexSize"].(int64); ok {
				mongoCollectionStats.WithLabelValues(dbName, collectionName, "index_size_bytes").Set(float64(indexSize))
			}

			log.Printf("Collected stats for collection %s.%s", dbName, collectionName)
		}
	}

	mongoTotalCollections.Set(float64(totalCollections))
	mongoTotalDocuments.Set(float64(totalDocuments))
	log.Printf("Total collections: %d, Total documents: %d", totalCollections, totalDocuments)
	log.Println("Collection metrics collection completed.")
}

func collectSystemMetrics() {
	log.Println("Collecting system metrics...")
	// Collect CPU usage
	cpuPercentages, _ := cpu.Percent(0, false)
	if len(cpuPercentages) > 0 {
		cpuUsage.Set(cpuPercentages[0])
		log.Printf("CPU Usage: %f%%", cpuPercentages[0])
	}

	// Collect Memory usage
	memStats, _ := mem.VirtualMemory()
	memoryUsage.Set(float64(memStats.Used))
	log.Printf("Memory Usage: %d bytes", memStats.Used)

	// Collect Disk usage
	if diskStat, err := disk.Usage("/"); err == nil {
		diskUsage.Set(float64(diskStat.Used))
		log.Printf("Disk Usage: %d bytes", diskStat.Used)
	}
	log.Println("System metrics collection completed.")
}

func collectAdditionalMetrics(client *mongo.Client, ctx context.Context) {
	log.Println("Collecting additional metrics...")
	serverStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err != nil {
		log.Printf("Failed to get server status for additional metrics: %v", err)
		mongoErrorCount.WithLabelValues("additional_metrics").Inc()
		return
	}

	// Cache Evictions
	if wiredTiger, ok := serverStatus["wiredTiger"].(bson.M); ok {
		if cache, ok := wiredTiger["cache"].(bson.M); ok {
			if evictions, ok := cache["evicted"].(int64); ok {
				mongoCacheEvictions.Set(float64(evictions))
				log.Printf("MongoDB cache evictions: %d", evictions)
			} else {
				log.Println("Failed to parse cache evictions.")
			}
		}
	}

	// Index Access Patterns
	if indexCounters, ok := serverStatus["indexCounters"].(bson.M); ok {
		if accesses, ok := indexCounters["accesses"].(bson.M); ok {
			for index, count := range accesses {
				if cnt, ok := count.(int64); ok {
					mongoIndexAccessPatterns.WithLabelValues(index).Set(float64(cnt))
					log.Printf("MongoDB index access pattern for %s: %d", index, cnt)
				}
			}
		}
	}

	// Client Connections
	if connections, ok := serverStatus["connections"].(bson.M); ok {
		if active, ok := connections["current"].(int64); ok {
			mongoActiveClientConnections.Set(float64(active))
			log.Printf("MongoDB active client connections: %d", active)
		}
		if total, ok := connections["totalCreated"].(int64); ok {
			mongoTotalClientConnections.Set(float64(total))
			log.Printf("MongoDB total client connections: %d", total)
		}
	}
	log.Println("Additional metrics collection completed.")
}

func main() {
	// Configure logging to include timestamps
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Initializing MongoDB Exporter...")
	var MongoDefaultPort string = "1203"
	versionFlag := flag.Bool("version", false, "Prints version information")
	flag.BoolVar(versionFlag, "V", false, "Prints version information (alias)")
	uri := flag.String("uri", os.Getenv("MONGO_URI"), "MongoDB URI")
	user := flag.String("user", os.Getenv("MONGO_USER"), "MongoDB username")
	password := flag.String("password", os.Getenv("MONGO_PASSWORD"), "MongoDB password")
	port := flag.String("port", MongoDefaultPort, "Port to serve metrics on, default is 1203")
	interval := flag.Duration("interval", 10*time.Second, "Interval for collecting metrics, default is 10s")

	// Flags to enable/disable specific metrics
	flag.BoolVar(&disableDBListing, "disable_db_listing", false, "Disable database listing metric")
	flag.BoolVar(&disableOpCounters, "disable_op_counters", false, "Disable operation counters metric")
	flag.BoolVar(&disableReplication, "disable_replication", false, "Disable replication metrics")
	flag.BoolVar(&disableMemoryUsage, "disable_memory_usage", false, "Disable memory usage metrics")
	flag.BoolVar(&disableLockMetrics, "disable_lock_metrics", false, "Disable lock metrics")
	flag.BoolVar(&disableNetworkMetrics, "disable_network_metrics", false, "Disable network metrics")

	// Help page to run binary

	if flag.Arg(0) == "help" || flag.Arg(0) == "-h" || flag.Arg(0) == "--help" {
		flag.PrintDefaults()
		log.Println("Custom MongoDB Exporter Help:")
		log.Println("This exporter collects metrics from MongoDB and exposes them in a format compatible with Prometheus.")
		log.Println("Flags:")
		log.Println("-disable_db_listing: Disable database listing metric")
		log.Println("-disable_lock_metrics: Disable lock metrics")
		log.Println("-disable_memory_usage: Disable memory usage metrics")
		log.Println("-disable_network_metrics: Disable network metrics")
		log.Println("-disable_op_counters: Disable operation counters metric")
		log.Println("-password: MongoDB password (default empty)")
		log.Println("-port: Port to serve metrics on (default 1203)")
		log.Println("-uri: MongoDB URI (required)")
		log.Println("-user: MongoDB username (default empty)")
		log.Println("-interval: Interval for collecting metrics (default 10s)")
		os.Exit(0)
	}

	flag.Parse()

	// Set the collection interval
	collectionInterval = *interval

	// Check if the version flag is set
	if *versionFlag {
		fmt.Printf("Custom Mongo Exporter Version(CME-1203): %s\n", version)
		return
	}

	if *uri == "" {
		log.Fatal("MongoDB URI is required. Set it using the --uri flag or the MONGO_URI environment variable.")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		for {
			collectMongoMetrics(*uri, *user, *password)
			// Context with timeout to prevent hanging operations
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			clientOpts := options.Client().ApplyURI(*uri)
			if *user != "" && *password != "" {
				clientOpts.SetAuth(options.Credential{Username: *user, Password: *password})
			}
			client, err := mongo.Connect(ctx, clientOpts)
			if err != nil {
				log.Printf("Failed to connect to MongoDB: %v", err)
				return
			}
			collectShardMetrics(client, ctx)
			collectReplicaSetMetrics(client, ctx)
			collectDetailedOperationMetrics(client, ctx)
			collectAdvancedMetrics(client, ctx)   // New advanced metrics collection
			collectCollectionMetrics(client, ctx) // New collection metrics
			collectAdditionalMetrics(client, ctx) // New additional metrics
			collectSystemMetrics()
			time.Sleep(collectionInterval)
		}
	}()

	log.Printf("Starting MongoDB Exporter on: %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

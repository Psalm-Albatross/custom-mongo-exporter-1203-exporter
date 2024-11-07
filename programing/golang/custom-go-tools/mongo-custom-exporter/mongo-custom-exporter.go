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

	// Flags for Optional Metrics
	disableDBListing      bool
	disableOpCounters     bool
	disableReplication    bool
	disableMemoryUsage    bool
	disableLockMetrics    bool
	disableNetworkMetrics bool
	// Version variable to check build version of custom-mongo-exporter tools
	version string
)

func init() {
	// Register all metrics
	prometheus.MustRegister(
		mongoUp, mongoDBCount, mongoDBList, mongoIdleConnections, mongoOpCounters,
		mongoReplicationLag, mongoMemoryUsage, mongoCacheUsage, mongoPageFaults,
		mongoLockTotalTime, mongoLockRatio, cpuUsage, memoryUsage, diskUsage,
		mongoNetworkIn, mongoNetworkOut, mongoCurrentConnections,
	)
}

func collectMongoMetrics(uri, user, password string) {
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
		return
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Check the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Printf("MongoDB ping failed: %v", err)
		mongoUp.Set(0)
		return
	}
	log.Println("Connected to MongoDB successfully!")
	mongoUp.Set(1)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("MongoDB ping failed: %v", err)
		mongoUp.Set(0)
		return
	}
	mongoUp.Set(1)

	serverStatus := bson.M{}
	if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err == nil {
		// Op Counters
		if !disableOpCounters {
			if opCounters, ok := serverStatus["opcounters"].(bson.M); ok {
				for opType, count := range opCounters {
					if cnt, ok := count.(int32); ok {
						mongoOpCounters.WithLabelValues(opType).Set(float64(cnt))
					}
				}
				log.Println("MongoDB operation counters collected.")
			}
		}

		// Memory Usage
		if !disableMemoryUsage {
			if memInfo, ok := serverStatus["mem"].(bson.M); ok {
				if resident, ok := memInfo["resident"].(int32); ok {
					mongoMemoryUsage.Set(float64(resident) * 1024 * 1024)
					log.Printf("MongoDB resident memory: %d MB", resident)
				}
				if faults, ok := memInfo["page_faults"].(int32); ok {
					mongoPageFaults.Set(float64(faults))
					log.Printf("MongoDB page faults: %d", faults)
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
	}

	// DB Count and Listing
	if !disableDBListing {
		dbs, err := client.ListDatabaseNames(ctx, bson.M{})
		if err != nil {
			log.Printf("Failed to list databases: %v", err)
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
		if err := client.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1}).Decode(&replStatus); err == nil {
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
		if err := client.Database("admin").RunCommand(ctx, bson.M{"serverStatus": 1}).Decode(&serverStatus); err == nil {
			if connections, ok := serverStatus["connections"].(bson.M); ok {
				if idle, ok := connections["available"].(int32); ok {
					mongoIdleConnections.Set(float64(idle))
					log.Printf("Idle connections: %d", idle)
				}
			}
		} else {
			log.Printf("Failed to collect idle connections: %v", err)
		}
	}

	// Collect Networking Metrics
	if !disableNetworkMetrics {
		networkStats, err := net.IOCounters(true)
		if err != nil {
			log.Printf("Failed to get network stats: %v", err)
			return
		}
		for _, netStat := range networkStats {
			mongoNetworkIn.Set(float64(netStat.BytesRecv))
			mongoNetworkOut.Set(float64(netStat.BytesSent))
			mongoCurrentConnections.Set(float64(netStat.PacketsSent + netStat.PacketsRecv)) // A simple estimate of current connections
			log.Printf("MongoDB network in: %d bytes, network out: %d bytes", netStat.BytesRecv, netStat.BytesSent)
		}
	}
}

func collectSystemMetrics() {
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
}

func main() {
	var MongoDefaultPort string = "1203"
	versionFlag := flag.Bool("version", false, "Prints version information")
	flag.BoolVar(versionFlag, "V", false, "Prints version information (alias)")
	uri := flag.String("uri", os.Getenv("MONGO_URI"), "MongoDB URI")
	user := flag.String("user", os.Getenv("MONGO_USER"), "MongoDB username")
	password := flag.String("password", os.Getenv("MONGO_PASSWORD"), "MongoDB password")
	port := flag.String("port", MongoDefaultPort, "Port to serve metrics on, default is 1203")

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
		os.Exit(0)
	}

	flag.Parse()

	// Check if the version flag is set
	if *versionFlag {
		fmt.Printf("Custom Mongo Exporter Version: %s\n", version)
		return
	}

	if *uri == "" {
		log.Fatal("MongoDB URI is required. Set it using the --uri flag or the MONGO_URI environment variable.")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		for {
			collectMongoMetrics(*uri, *user, *password)
			collectSystemMetrics()
			time.Sleep(10 * time.Second)
		}
	}()

	log.Printf("Starting MongoDB Exporter on: %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

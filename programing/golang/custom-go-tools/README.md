# Custom MongoDB Exporter

This custom MongoDB exporter collects metrics from MongoDB and exposes them in a format compatible with Prometheus. It provides detailed insights into MongoDB performance, health, and operations.

## Features

- MongoDB up and reachable status
- Total number of databases in MongoDB
- List of databases in MongoDB
- Number of idle connections to MongoDB
- MongoDB operation counters by type (insert, query, update, delete)
- Replication lag for each MongoDB node
- MongoDB memory usage in bytes
- MongoDB cache usage in bytes
- MongoDB page faults
- Total lock time for read and write operations
- Lock ratio for read and write operations
- System CPU usage percentage
- System memory usage in bytes
- System disk usage in bytes
- Total bytes received and sent by MongoDB
- Current number of connections to MongoDB
- MongoDB transactions per second
- MongoDB query execution time and latency
- MongoDB query performance metrics
- MongoDB index usage
- Number of documents inserted, updated, and deleted
- Number of active and total connections to MongoDB
- MongoDB operation latency
- Total bytes read from and written to disk by MongoDB
- Total number of MongoDB errors encountered

## Usage

### Building the Exporter

You can build the exporter using the provided `Makefile` or `build.sh` script.

#### Using Makefile

```sh
make build
```

This will build the binaries for multiple operating systems and architectures and place them in the `bin` directory.

#### Using build.sh

```sh
cd scripts
./build.sh
```

This will build the binaries for Windows, Linux, and macOS and place them in the `bin` directory.

### Running the Exporter

To run the exporter, use the following command:

```sh
./bin/custom-1203-mongo-exporter-linux-amd64 --uri <MONGO_URI> --user <MONGO_USER> --password <MONGO_PASSWORD> --port <PORT>
```

Replace `<MONGO_URI>`, `<MONGO_USER>`, `<MONGO_PASSWORD>`, and `<PORT>` with your MongoDB URI, username, password, and the port to serve metrics on, respectively.

### Flags

- `--uri`: MongoDB URI (required)
- `--user`: MongoDB username (default empty)
- `--password`: MongoDB password (default empty)
- `--port`: Port to serve metrics on (default 1203)
- `--interval`: Interval for collecting metrics (default 10s)
- `--disable_db_listing`: Disable database listing metric
- `--disable_op_counters`: Disable operation counters metric
- `--disable_replication`: Disable replication metrics
- `--disable_memory_usage`: Disable memory usage metrics
- `--disable_lock_metrics`: Disable lock metrics
- `--disable_network_metrics`: Disable network metrics

### Example

```sh
./bin/custom-1203-mongo-exporter-linux-amd64 --uri mongodb://localhost:27017 --user admin --password secret --port 1203
```

This will start the exporter and serve metrics on port 1203.

### Prometheus Configuration

Add the following job to your Prometheus configuration to scrape metrics from the exporter:

```yaml
scrape_configs:
  - job_name: 'custom-mongo-exporter'
    static_configs:
      - targets: ['localhost:1203']
```

### Grafana Dashboard

You can use the provided Grafana dashboard JSON file to visualize the metrics collected by the exporter. Import the `mongo_grafana_dashboard.json` file into your Grafana instance to get started.

### Alerts

Prometheus alerting rules are provided in the `critical_mongo_alerting.promql` file. These rules cover various critical conditions such as MongoDB being down, high replication lag, high memory usage, slow queries, and more. Import these rules into your Prometheus configuration to enable alerting.

### Exposed Metrics

A list of exposed metrics is provided in the `exposed_mongo_metrics.promql` file. This file includes examples of all the metrics collected by the exporter.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License.

# Custom MongoDB Exporter

Custom 1203 MongoDB (Custom-1203-Mongo-Exporter) Exporter is a tool designed to collect and expose MongoDB metrics in a format compatible with Prometheus. The exporter provides detailed metrics for MongoDB, including database statistics, replication lag, memory usage, and more.

---

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Flags](#flags)
- [Example Commands](#example-commands)
- [Metrics Exposed](#metrics-exposed)
- [Building From Source](#building-from-source)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Installation

### Download Binary
You can download the latest release of `custom-1203-mongo-exporter` from the [GitHub Releases](https://github.com/Psalm-Albatross/custom-mongo-exporter-1203-exporter//releases) page. Choose the appropriate binary for your platform and architecture.

### Example Naming for Binaries
- **Linux x86_64**: `custom-1203-mongo-exporter-v1.0.0-linux-amd64`
- **Windows x86_64**: `custom-1203-mongo-exporter-v1.0.0-windows-amd64.exe`
- **macOS ARM64**: `custom-1203-mongo-exporter-v1.0.0-darwin-arm64`

---

## Usage

### Running the Exporter
To run the exporter, execute the binary from the command line with the necessary flags.

```bash
./custom-1203-mongo-exporter --uri "mongodb://localhost:27017" --port 1203
```

### Required Flags

  -	--uri: The MongoDB URI (e.g., mongodb://localhost:27017).
  - --port: The port on which the exporter will serve metrics (default: 1203).

## Flags


### Example commands

Connect to a Local MongoDB Instance

```sh
./custom-1203-mongo-exporter --uri "mongodb://localhost:27017"
```

Connect to a MongoDB Instance with Authentication

```sh
./custom-1203-mongo-exporter --uri "mongodb://localhost:27017" --user "myUser" --password "myPassword"
```

### Customize Metrics Collection

```sh
./custom-1203-mongo-exporter --uri "mongodb://localhost:27017" --disable_op_counters --disable_memory_usage
```


### Metrics Exposed

The exporter provides several metrics, including:

- **MongoDB Health Metrics**
  - mongodb_up: Indicates if MongoDB is reachable.
  - mongodb_db_count: Total number of databases.

- **System Resource Metrics**
  - system_cpu_usage_percent: CPU usage percentage.
  - system_memory_usage_bytes: Memory usage in bytes.

- **Database Operation Metrics**
  - mongodb_op_counters: Operation counters (insert, query, update, delete).
  - mongodb_memory_usage_bytes: Memory usage in bytes.

- **Replication Metrics**
  - mongodb_replication_lag: Replication lag for each node.

- **Network Metrics**
  - mongodb_network_in_bytes: Total bytes received by MongoDB.
  - mongodb_network_out_bytes: Total bytes sent by MongoDB.



## Acknowledgments

This project is inspired by the mongo-exporter project, which provides Prometheus metrics for MongoDB. We have built upon the ideas and concepts of the mongo-exporter to create a custom version that suits specific needs for the use of internal use cases for the organization.
                                                                                                                       
***- Psalm Albatross***
                                              
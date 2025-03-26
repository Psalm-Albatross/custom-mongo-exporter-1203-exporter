# Release Notes - v2.0.0

## New Features
- **Enhanced Metrics Collection**: Added support for collecting additional MongoDB metrics, including:
  - Query execution time
  - Connection pool statistics
  - Index usage statistics
- **Total Collection Count**: Added a metric to track the total number of collections in the database.
- **Customizable Exporter Configuration**: Introduced a configuration file to allow users to customize the metrics they want to export.
- **Improved Logging**: Added structured logging with support for different log levels (e.g., DEBUG, INFO, ERROR).

## Functionality Additions
- **Prometheus Integration**: Improved Prometheus scraping compatibility with optimized metric naming conventions.
- **Authentication Support**: Added support for MongoDB authentication mechanisms, including SCRAM-SHA-1 and SCRAM-SHA-256.
- **TLS/SSL Support**: Enabled secure connections to MongoDB instances using TLS/SSL.

## Improvements
- **Performance Optimization**: Reduced memory usage and improved the efficiency of metric collection.
- **Error Handling**: Enhanced error handling with detailed error messages and retry mechanisms for transient failures.
- **Documentation**: Updated documentation with detailed setup instructions and examples for using the new features.

## Breaking Changes
- Metric names have been updated to follow Prometheus best practices. Users upgrading from previous versions may need to update their Prometheus queries.

## Upgrade Instructions
1. Update your configuration file to include the new options for metrics and authentication.
2. Review and update your Prometheus queries to align with the new metric naming conventions.
3. Restart the exporter with the updated configuration.

For more details, refer to the [documentation](./doca/release_doc.md).

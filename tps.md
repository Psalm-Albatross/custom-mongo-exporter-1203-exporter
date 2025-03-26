### Explanation of TPS (Transactions Per Second) in MongoDB

**Transactions Per Second (TPS)** is a metric that measures the number of transactions (or operations) executed by MongoDB per second. In MongoDB, transactions can include operations like inserts, updates, deletes, and queries. TPS is critical for understanding the throughput and performance of the database.

### How TPS is Calculated
1. MongoDB provides operation counters (`opcounters`) in the `serverStatus` command. These counters include:
   - `insert`: Number of insert operations.
   - `update`: Number of update operations.
   - `delete`: Number of delete operations.
   - `query`: Number of query operations.
   - `command`: Number of command operations.

2. To calculate TPS, you need to:
   - Fetch the current values of these counters.
   - Subtract the previous values from the current values to get the number of operations performed during the interval.
   - Divide the difference by the interval duration (in seconds) to get the TPS.

### Why TPS Might Not Show Data
1. **Incorrect Calculation**: If the difference between the current and previous counters is not calculated, TPS will always show `0`.
2. **No Activity**: If there are no operations happening in the database, the counters will not increment, and TPS will remain `0`.
3. **Prometheus Metric Not Updated**: If the Prometheus metric for TPS is not updated correctly, it will not show data in Grafana or Prometheus.
# PATH Metrics Patterns

## Current Prometheus Metrics Setup

### Structure
- **Central metrics files**:
  - `metrics/gateway.go` - Gateway-level metrics (requests, duration, response size)
  - `metrics/server.go` - Prometheus metrics server setup
  - `metrics/prometheus_reporter.go` - Metrics reporter interface
  - `cmd/metrics.go` - Metrics server initialization

### Protocol-specific metrics
- `metrics/protocol/shannon/metrics.go` - Shannon protocol metrics
- `metrics/protocol/morse/metrics.go` - Morse protocol metrics

### QoS-specific metrics
- `metrics/qos/evm/metrics.go` - EVM QoS metrics
- `metrics/qos/solana/metrics.go` - Solana QoS metrics

## Pattern Analysis

### Common Constants
- All metrics files use `pathProcess = "path"` as the subsystem name
- Each file defines specific metric name constants (e.g., `requestsTotal`, `relaysTotalMetric`)

### Registration Pattern
- Each metrics file has an `init()` function that calls `prometheus.MustRegister()` for its metrics
- Metrics are defined as package-level variables using `prometheus.NewCounterVec`, `prometheus.NewHistogramVec`

### Metric Definition Structure
```go
var myMetric = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Subsystem: pathProcess,  // Always "path"
        Name:      metricName,   // Defined as constant
        Help:      "Description...",
    },
    []string{"label1", "label2"},  // Labels
)
```

### Publishing Pattern
- Each metrics file has a `PublishMetrics()` function
- Uses `prometheus.Labels{}` to set label values
- Calls `.With(labels).Inc()` or `.Observe(value)` on metrics

## Version Handling
- Project uses git-based versioning: `git describe --tags --always --dirty`
- No existing version metric found in current setup
- Build process uses ldflags in `makefiles/release.mk`
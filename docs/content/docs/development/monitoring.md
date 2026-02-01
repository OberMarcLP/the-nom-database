+++
title = "Monitoring"
description = "Logging and metrics"
weight = 450
icon = "monitor"
toc = true
+++

# Monitoring & Logging Guide

## Overview

The Nom Database implements comprehensive monitoring and structured logging using zerolog and custom metrics collection. This guide covers how to use and interpret the monitoring features.

## Structured Logging

### Log Format

The application uses **zerolog** for high-performance structured logging with the following features:

- **Structured Fields**: All log entries include structured key-value pairs
- **Request Correlation**: Unique request IDs for tracing requests
- **Performance Metrics**: Request duration, response size, status codes
- **Color-Coded Console**: Easy-to-read colored output in development
- **JSON Output**: Structured JSON logs for production (set `LOG_FORMAT=json`)

### Log Levels

- **DEBUG**: Detailed information for debugging (enabled with `DEBUG=true`)
- **INFO**: General informational messages
- **WARN**: Warning messages for non-critical issues
- **ERROR**: Error messages for failures
- **FATAL**: Critical errors that cause application exit

### Log Output Examples

#### Console Format (Development)
```
2025-12-30 20:13:44 INF HTTP request completed bytes=1124 duration=4.333917 ip=192.168.65.1:55864 method=GET path=/api/categories request_id=62d30079-6774-46f6-b623-83680479d9a7 status=200
```

#### JSON Format (Production)
```json
{
  "level": "info",
  "time": "2025-12-30T20:13:44Z",
  "message": "HTTP request completed",
  "bytes": 1124,
  "duration": 4.333917,
  "ip": "192.168.65.1:55864",
  "method": "GET",
  "path": "/api/categories",
  "request_id": "62d30079-6774-46f6-b623-83680479d9a7",
  "status": 200
}
```

### Structured Log Fields

Each HTTP request log includes:

| Field | Description | Example |
|-------|-------------|---------|
| `request_id` | Unique UUID for request tracing | `62d30079-6774-46f6-b623-83680479d9a7` |
| `method` | HTTP method | `GET` |
| `path` | Request path | `/api/categories` |
| `ip` | Client IP address | `192.168.65.1:55864` |
| `duration` | Request processing time | `4.333917` (milliseconds) |
| `status` | HTTP status code | `200` |
| `bytes` | Response size in bytes | `1124` |

### Configuring Logging

#### Development Mode

Enable debug logging:
```bash
docker compose up -d
# or
DEBUG=true go run ./cmd/server
```

#### Production Mode

Use JSON format for log aggregation:
```bash
LOG_FORMAT=json go run ./cmd/server
```

Or in docker-compose.yml:
```yaml
environment:
  LOG_FORMAT: json
  DEBUG: false
```

## Request Correlation IDs

Every request is assigned a unique UUID that appears in:

1. **Response Headers**: `X-Request-ID` header
2. **Log Entries**: `request_id` field
3. **Error Messages**: Contextual error logging

### Using Request IDs

#### From Client
```bash
# Server generates ID automatically
curl http://localhost:8080/api/restaurants -v

# Or provide your own
curl -H "X-Request-ID: my-custom-id" http://localhost:8080/api/restaurants
```

#### Tracing Requests
```bash
# Filter logs by request ID
docker compose logs backend | grep "62d30079-6774-46f6-b623-83680479d9a7"
```

## Metrics Collection

### Real-Time Metrics

Access current metrics at: **http://localhost:8080/api/metrics**

### Metrics Endpoint

```bash
curl http://localhost:8080/api/metrics
```

Response:
```json
{
  "total_requests": 1234,
  "total_errors": 5,
  "requests_by_method": {
    "GET": 800,
    "POST": 300,
    "PUT": 100,
    "DELETE": 34
  },
  "requests_by_path": {
    "/api/restaurants": 450,
    "/api/categories": 200,
    "/api/ratings": 150
  },
  "requests_by_status": {
    "200": 1100,
    "201": 50,
    "400": 20,
    "404": 50,
    "500": 14
  },
  "avg_response_time": "2.5ms",
  "p50_response_time": "1.2ms",
  "p95_response_time": "8.5ms",
  "p99_response_time": "15.3ms",
  "uptime": "2h15m30s"
}
```

### Metrics Explained

| Metric | Description |
|--------|-------------|
| `total_requests` | Total number of HTTP requests processed |
| `total_errors` | Number of 5xx server errors |
| `requests_by_method` | Request count by HTTP method |
| `requests_by_path` | Request count by API path |
| `requests_by_status` | Request count by HTTP status code |
| `avg_response_time` | Average request duration |
| `p50_response_time` | 50th percentile (median) response time |
| `p95_response_time` | 95th percentile response time |
| `p99_response_time` | 99th percentile response time |
| `uptime` | Time since metrics collection started |

### Periodic Metrics Logging

Metrics are automatically logged every **5 minutes** with structured fields:

```
INFO 📊 Metrics Summary total_requests=1234 total_errors=5 avg_response_time=2.5ms ...
```

## Performance Monitoring

### Response Time Percentiles

- **p50 (Median)**: Half of requests are faster than this
- **p95**: 95% of requests are faster than this (target for SLA)
- **p99**: 99% of requests are faster than this (outlier detection)

### Interpreting Metrics

#### Healthy Application
```json
{
  "avg_response_time": "2.5ms",
  "p50_response_time": "1.2ms",
  "p95_response_time": "8.5ms",
  "p99_response_time": "15.3ms"
}
```

#### Performance Issues
```json
{
  "avg_response_time": "150ms",
  "p50_response_time": "80ms",
  "p95_response_time": "500ms",
  "p99_response_time": "2s"
}
```
*Action: Investigate database queries, add caching, optimize handlers*

#### High Error Rate
```json
{
  "total_requests": 1000,
  "total_errors": 150,
  "requests_by_status": {
    "500": 150
  }
}
```
*Action: Check logs for error details, investigate failing endpoints*

## Viewing Logs

### Docker Compose

```bash
# View all logs
docker compose logs backend

# Follow logs in real-time
docker compose logs -f backend

# Last 100 lines
docker compose logs --tail=100 backend

# Filter by log level
docker compose logs backend | grep "ERR\|FATAL"

# Filter by request ID
docker compose logs backend | grep "request_id=abc123"

# Filter by endpoint
docker compose logs backend | grep "/api/restaurants"
```

### Log Analysis

#### Find slow requests (>100ms)
```bash
docker compose logs backend | grep "duration=" | awk -F'duration=' '{print $2}' | sort -n
```

#### Count requests by status code
```bash
docker compose logs backend | grep "status=" | awk -F'status=' '{print $2}' | cut -d' ' -f1 | sort | uniq -c
```

#### Find errors
```bash
docker compose logs backend | grep "ERR"
```

## Production Recommendations

### Log Aggregation

For production, use a log aggregation service:

1. **JSON Format**: Set `LOG_FORMAT=json`
2. **Shipping**: Use Fluentd, Filebeat, or CloudWatch agent
3. **Storage**: Send to Elasticsearch, Splunk, or CloudWatch Logs
4. **Analysis**: Use Kibana, Splunk, or CloudWatch Insights

### Example: CloudWatch Logs

```yaml
# docker-compose.yml
services:
  backend:
    logging:
      driver: awslogs
      options:
        awslogs-region: us-east-1
        awslogs-group: nomdb-backend
        awslogs-stream: backend
```

### Example: Elasticsearch + Filebeat

```yaml
# filebeat.yml
filebeat.inputs:
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'

processors:
  - add_docker_metadata: ~
  - decode_json_fields:
      fields: ["message"]
      target: ""

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

### Alerting

Set up alerts based on metrics:

1. **High Error Rate**: `total_errors / total_requests > 0.05` (5% error rate)
2. **Slow Response**: `p95_response_time > 500ms`
3. **High Request Count**: `requests_by_path["/api/endpoint"] > 1000/min` (potential DoS)

### Monitoring Dashboards

Create dashboards with:
- Request rate over time
- Error rate trends
- Response time percentiles (p50, p95, p99)
- Top endpoints by request count
- Top endpoints by errors
- Geographic distribution of requests

## Debugging with Logs

### Example: Tracing a Slow Request

1. Make a request and capture the request ID from headers:
```bash
curl -v http://localhost:8080/api/restaurants | grep X-Request-ID
# X-Request-ID: abc123-def456-789
```

2. Find all logs for that request:
```bash
docker compose logs backend | grep "abc123-def456-789"
```

3. Analyze the logs:
- Check duration
- Look for database queries
- Identify bottlenecks

### Example: Debugging an Error

1. Find recent errors:
```bash
docker compose logs backend | grep "ERR" | tail -10
```

2. Look for stack traces and context
3. Check request_id to trace the full request lifecycle
4. Investigate related requests with similar patterns

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DEBUG` | `false` | Enable debug logging |
| `LOG_FORMAT` | `console` | Log format: `console` or `json` |
| `PORT` | `8080` | Server port |

## Best Practices

1. **Always include request_id** in error reports
2. **Monitor p95/p99** response times, not just averages
3. **Set up alerts** for error rates and slow responses
4. **Use JSON format** in production for log aggregation
5. **Keep logs** for at least 30 days for debugging
6. **Analyze metrics** regularly to identify trends
7. **Correlate logs** with metrics for full observability

## Troubleshooting

### No Logs Appearing

Check if debug mode is enabled:
```bash
docker compose logs backend | grep "Debug mode"
```

### Metrics Reset to Zero

Metrics are reset when the application restarts. For persistent metrics, integrate with Prometheus or similar.

### High Memory Usage from Logs

Logs are written to stdout/stderr. If memory is an issue:
1. Use log rotation
2. Ship logs to external service
3. Limit log retention

## Next Steps

For full production monitoring, consider:

1. **Prometheus Integration** - Time-series metrics
2. **Grafana Dashboards** - Visual monitoring
3. **Alert Manager** - Automated alerting
4. **Distributed Tracing** - OpenTelemetry/Jaeger
5. **Error Tracking** - Sentry integration

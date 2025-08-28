# TraceMonitor Collector

trace-monitor-collector is a lightweight Go service for collecting application traces in real-time over UDP and exposing them as Prometheus metrics or json.

It is designed to be used together with language-specific clients that send JSON-encoded UDP datagrams describing trace lifecycle events

## Configuration
Config example: config.yaml

## Build and Run
```
# Go 1.18+
go mod vendor
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/trace-monitor-collector .

./build/trace-monitor-collector -config ./config.yaml
```

## HTTP Endpoints
`/getall.json`: All traces data  
`/getall`: All traces data human-readable  
`/console/metrics`: Prometheus format metrics  

## UDP Protocol
### init-trace

Sent once when a new trace is opened.
```json
{
    "method": "init-trace",
    "sentAt": "2025-08-28T16:34:12.123456+03:00",
    "pid": "12345",
    "traceId": "abc123",
    "data": {
    "context": { "userId": "42" },         
    "tags": { "service": "api" },          
        "openedAt": "2025-08-28T16:34:12.123456+03:00"
    }
}
```

### set-trace-current-span

Updates the current span of a trace (optionally with its parents).

```json
{
    "method": "set-trace-current-span",
    "sentAt": "2025-08-28T16:34:15.000000+03:00",
    "pid": "12345",
    "traceId": "abc123",
    "data": {
        "span": {
            "id": "span-1",
            "parent": null,                               
            "openedAt": "2025-08-28T16:34:15.000000+03:00",
            "name": "db.query",
            "context": { "sql": "SELECT ..." },
            "tags": { "status": "ok" },
            "debugBacktrace": "..."                       
        },
        "parentSpans": [
            { "id": "span-0", "parent": null, "...": "..." }
        ]
    }
}
```

Notes:

Number of parent spans included is limited by countParentForDataPackage.

parentSpans are serialized same as span.

### free-pid

Signals that a process has finished using a trace.

```json
{
    "method": "free-pid",
    "sentAt": "2025-08-28T16:34:30.000000+03:00",
    "pid": "12345",
    "traceId": "abc123",
    "data": null
}
```



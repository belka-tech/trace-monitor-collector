# Prometheus metrics

[Back to README](../README.md#http--metrics)

The Collector serves metrics at `/console/metrics` on its configured `http_addr`. This reference covers its eight `trace_monitor_*` metrics.

## Summary

| Metric | Type | Unit | Meaning |
| --- | --- | --- | --- |
| [trace_monitor_count_active_pid](#trace_monitor_count_active_pid) | Gauge | PIDs | PID entries currently held in memory. |
| [trace_monitor_total_trace_set](#trace_monitor_total_trace_set) | Counter | commands | `init-trace` handler calls. |
| [trace_monitor_total_span_set](#trace_monitor_total_span_set) | Counter | commands | Current-span updates with non-null data. |
| [trace_monitor_total_all_span_close](#trace_monitor_total_all_span_close) | Counter | commands | Current-span clear commands (`data: null`). |
| [trace_monitor_total_trace_delete](#trace_monitor_total_trace_delete) | Counter | commands | `free-pid` handler calls. |
| [trace_monitor_total_packages_caught](#trace_monitor_total_packages_caught) | Counter | datagrams | Datagrams read by the UDP listener. |
| [trace_monitor_total_packages_parse](#trace_monitor_total_packages_parse) | Counter | datagrams | Packets that reach the end of command processing. |
| [trace_monitor_total_channel_reset](#trace_monitor_total_channel_reset) | Counter | queue resets | Overflow-triggered resets of UDP packet queues. |

## Scraping and labels

Configure Prometheus to scrape each node's Collector:

```yaml
scrape_configs:
  - job_name: trace-monitor
    metrics_path: /console/metrics
    static_configs:
      - targets: ["app-01:20000", "app-02:20000"]
```

Replace the example hostnames with addresses reachable from Prometheus. All queries below use this `job` name.

The eight `trace_monitor_*` metrics have these labels:

| Label | Source | Example |
| --- | --- | --- |
| `node` | Collector hostname, truncated at the first dot. | `app-01` |
| `app` | `app_name` in the Collector configuration. | `api` |
| `env` | `env` in the Collector configuration. | `production` |

These labels describe the Collector instance. Trace IDs, PIDs, span names and client-supplied tags are not exported as metric labels. Prometheus adds target labels such as `job` and `instance` when scraping.

Counters accumulate within the Collector process and start again at zero after a restart. Use [rate()](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate) for a per-second average or [increase()](https://prometheus.io/docs/prometheus/latest/querying/functions/#increase) for an estimated increase over a time window. Both handle counter resets; `increase()` may return fractional values because it extrapolates between scrapes. Apply `rate()` before summing across instances. Read gauges directly or use gauge-appropriate functions such as `max_over_time()`.

The descriptions below follow the actual update paths in [prometheus.go](../prometheus.go), [udpListener.go](../udpListener.go) and [traceCollection.go](../traceCollection/traceCollection.go). In particular, the four trace/span command counters increment before validation, so they measure handler calls, including rejected and no-op commands.

## Collector metrics

### `trace_monitor_count_active_pid`

Gauge · PIDs.

Counts the PID entries currently tracked by the Collector. A new entry can be created by either `init-trace` or a span update for an unknown PID. Nested spans still occupy the same PID slot.

The value decreases when an entry is removed by `free-pid`, PHP-FPM cleanup or another state-replacement path. Stale entries can remain when a client disappears without cleanup. Read this gauge directly: subtracting the trace command counters does not reproduce the current number of entries.

Current tracked PIDs per Collector:

```promql
trace_monitor_count_active_pid{job="trace-monitor"}
```

### `trace_monitor_total_trace_set`

Counter · commands.

Increments when an `init-trace` command reaches its handler, before checking the existing state and message chronology. Repeated initialization of the same trace and rejected initialization attempts contribute to the counter.

Use it to follow incoming trace-initialization activity. It cannot establish the number of unique traces or successful application requests.

Trace-initialization commands per second, averaged over five minutes:

```promql
rate(trace_monitor_total_trace_set{job="trace-monitor"}[5m])
```

### `trace_monitor_total_span_set`

Counter · commands.

Increments for `set-trace-current-span` commands whose `data` is non-null, before chronology validation. One command contributes one increment regardless of how many parent spans it carries.

The PHP client sends these updates both when opening a span and when closing a nested span restores its parent. Out-of-order updates also increment the counter. Updates with `data: null` go to `trace_monitor_total_all_span_close` instead.

Current-span update commands per second:

```promql
rate(trace_monitor_total_span_set{job="trace-monitor"}[5m])
```

### `trace_monitor_total_all_span_close`

Counter · commands.

Increments for `set-trace-current-span` commands with `data: null`. A successful command clears the stored current span while retaining the matching trace entry.

Despite the name, the value counts clear commands, not every individual span that closes. It also increases when the PID is already absent or the command is rejected as out of order. Closing a nested span that restores a parent is counted by `trace_monitor_total_span_set`.

Current-span clear commands per second:

```promql
rate(trace_monitor_total_all_span_close{job="trace-monitor"}[5m])
```

### `trace_monitor_total_trace_delete`

Counter · commands.

Increments when a `free-pid` command reaches its handler, before checking whether the PID exists or whether the timestamp is newer. Repeated releases and rejected releases therefore contribute to this counter.

Automatic removal by PHP-FPM and replacement of an existing PID's state do not increment it. Use it to monitor client release activity; it does not count all removals or prove that application work completed successfully.

Client release commands per second:

```promql
rate(trace_monitor_total_trace_delete{job="trace-monitor"}[5m])
```

### `trace_monitor_total_packages_caught`

Counter · datagrams.

Increments after a successful UDP read and its queue insertion or overflow handling. It includes malformed messages, unknown methods and packets that may subsequently be discarded by a queue reset.

This measures application-level UDP ingress across all configured ports. Packets lost before the process reads them are not counted, so the counter cannot measure total UDP delivery loss.

Received datagrams per second:

```promql
rate(trace_monitor_total_packages_caught{job="trace-monitor"}[5m])
```

### `trace_monitor_total_packages_parse`

Counter · datagrams.

Increments at the end of the packet-processing function. The current implementation continues after JSON parsing errors and command-handler errors, so unknown methods, malformed messages and rejected updates can still increase this counter. A recovered panic that exits processing before the increment does not.

Use it to follow processing throughput. The difference from `trace_monitor_total_packages_caught` can include queued, discarded or interrupted packets; it is not an exact queue length or validation-error count. The counters are also read separately during a scrape.

Packets reaching the end of processing per second:

```promql
rate(trace_monitor_total_packages_parse{job="trace-monitor"}[5m])
```

### `trace_monitor_total_channel_reset`

Counter · queue resets.

Increments when a per-port queue is full and the Collector drains pending packets before enqueueing the latest packet. The metric aggregates resets across all ports and has no port label.

A reset is one event, regardless of how many packets were discarded. Growth indicates queue pressure: inspect ingress and processing rates, Collector resource usage and the `packets_size` setting. A zero value only establishes that this queue-reset path has not run; network or socket-level packet loss can still occur.

Estimated queue resets in the last five minutes; a positive value identifies affected targets:

```promql
increase(trace_monitor_total_channel_reset{job="trace-monitor"}[5m]) > 0
```

## Standard metrics

The same endpoint also exposes standard [`go_*`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewGoCollector), [`process_*`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewProcessCollector) and [`promhttp_*`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentMetricHandler) metrics from the Prometheus Go client. These describe the Collector process and its metrics endpoint, not the applications sending traces. They do not inherit the Collector's `node`, `app` or `env` labels. The available series depend on the platform and the Go/client versions.

## Coverage and limits

Trace/span durations, request latency, UDP byte volume, current queue depth, exact packet-loss counts and command-validation errors have no dedicated metrics in this version. Use the live trace endpoints and logs when investigating individual traces or rejected commands. In particular, the packet and trace command counters cannot establish application success rates.

Prometheus also creates target-health series such as `up` and `scrape_duration_seconds` while scraping. These are generated by Prometheus and are not served by the Collector's `/console/metrics` endpoint.

# Prometheus metrics

[Back to README](../README.md#http--metrics)

The Collector serves metrics at `/console/metrics` on its configured `http_addr`. Eight metrics describe trace collection; the same endpoint includes standard Go runtime, process and scrape-handler metrics from the Prometheus client pinned in [go.mod](../go.mod).

## Summary

The table lists metric families. The GC summary also exposes `_sum` and `_count` series. Process metrics depend on platform support and access to process statistics; the process section below describes the Linux output.

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
| [promhttp_metric_handler_requests_total](#promhttp_metric_handler_requests_total) | Counter | HTTP requests | Completed metrics requests, grouped by HTTP status. |
| [promhttp_metric_handler_requests_in_flight](#promhttp_metric_handler_requests_in_flight) | Gauge | HTTP requests | Metrics requests currently being served. |
| [go_info](#go_info) | Gauge | constant 1 | Go version used to build the Collector. |
| [go_goroutines](#go_goroutines) | Gauge | goroutines | Goroutines currently present in the Collector. |
| [go_threads](#go_threads) | Gauge | OS threads | OS-thread count reported by the Go runtime. |
| [go_gc_duration_seconds](#go_gc_duration_seconds) | Summary | seconds | Garbage-collection pause statistics. |
| [go_memstats_alloc_bytes](#go_memstats_alloc_bytes) | Gauge | bytes | Heap memory occupied by allocated objects. |
| [go_memstats_alloc_bytes_total](#go_memstats_alloc_bytes_total) | Counter | bytes | Cumulative heap allocation volume. |
| [go_memstats_heap_alloc_bytes](#go_memstats_heap_alloc_bytes) | Gauge | bytes | Heap memory occupied by allocated objects. |
| [go_memstats_heap_sys_bytes](#go_memstats_heap_sys_bytes) | Gauge | bytes | Heap address space obtained by the Go runtime. |
| [go_memstats_heap_idle_bytes](#go_memstats_heap_idle_bytes) | Gauge | bytes | Heap spans currently available for reuse. |
| [go_memstats_heap_inuse_bytes](#go_memstats_heap_inuse_bytes) | Gauge | bytes | Heap spans assigned to object storage. |
| [go_memstats_heap_released_bytes](#go_memstats_heap_released_bytes) | Gauge | bytes | Idle heap memory returned to the OS. |
| [go_memstats_heap_objects](#go_memstats_heap_objects) | Gauge | objects | Currently allocated heap objects. |
| [go_memstats_mallocs_total](#go_memstats_mallocs_total) | Counter | allocations | Cumulative heap object allocations. |
| [go_memstats_frees_total](#go_memstats_frees_total) | Counter | frees | Cumulative heap object frees. |
| [go_memstats_lookups_total](#go_memstats_lookups_total) | Counter | lookups | Legacy pointer-lookup counter; currently zero. |
| [go_memstats_stack_inuse_bytes](#go_memstats_stack_inuse_bytes) | Gauge | bytes | Memory used by goroutine stacks. |
| [go_memstats_stack_sys_bytes](#go_memstats_stack_sys_bytes) | Gauge | bytes | Memory assigned to runtime stacks. |
| [go_memstats_mspan_inuse_bytes](#go_memstats_mspan_inuse_bytes) | Gauge | bytes | Active heap-span metadata. |
| [go_memstats_mspan_sys_bytes](#go_memstats_mspan_sys_bytes) | Gauge | bytes | Memory reserved for heap-span metadata. |
| [go_memstats_mcache_inuse_bytes](#go_memstats_mcache_inuse_bytes) | Gauge | bytes | Active allocation-cache metadata. |
| [go_memstats_mcache_sys_bytes](#go_memstats_mcache_sys_bytes) | Gauge | bytes | Memory reserved for allocation-cache metadata. |
| [go_memstats_buck_hash_sys_bytes](#go_memstats_buck_hash_sys_bytes) | Gauge | bytes | Profiling bucket metadata. |
| [go_memstats_gc_sys_bytes](#go_memstats_gc_sys_bytes) | Gauge | bytes | Runtime metadata in the legacy GC field. |
| [go_memstats_other_sys_bytes](#go_memstats_other_sys_bytes) | Gauge | bytes | Other Go runtime memory. |
| [go_memstats_sys_bytes](#go_memstats_sys_bytes) | Gauge | bytes | Total memory obtained by the Go runtime. |
| [go_memstats_next_gc_bytes](#go_memstats_next_gc_bytes) | Gauge | bytes | Current heap-size goal for garbage collection. |
| [go_memstats_last_gc_time_seconds](#go_memstats_last_gc_time_seconds) | Gauge | Unix seconds | Time of the latest garbage collection. |
| [process_cpu_seconds_total](#process_cpu_seconds_total) | Counter | CPU seconds | Collector process CPU time. |
| [process_resident_memory_bytes](#process_resident_memory_bytes) | Gauge | bytes | Resident memory of the Collector process. |
| [process_virtual_memory_bytes](#process_virtual_memory_bytes) | Gauge | bytes | Virtual address space of the Collector. |
| [process_virtual_memory_max_bytes](#process_virtual_memory_max_bytes) | Gauge | bytes | Process address-space limit. |
| [process_open_fds](#process_open_fds) | Gauge | file descriptors | Open descriptors in the Collector process. |
| [process_max_fds](#process_max_fds) | Gauge | file descriptors | Process file-descriptor limit. |
| [process_start_time_seconds](#process_start_time_seconds) | Gauge | Unix seconds | Collector process start time. |

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

These labels describe the Collector instance. Trace IDs, PIDs, span names and client-supplied tags are not exported as metric labels. Standard metrics do not inherit `node`, `app` or `env`; their extra labels are documented below. Prometheus adds target labels such as `job` and `instance` when scraping.

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

## Scrape handler metrics

These metrics are registered by the default [Prometheus HTTP handler](https://github.com/prometheus/client_golang/blob/v1.14.0/prometheus/promhttp/http.go). They describe access to the metrics endpoint.

### `promhttp_metric_handler_requests_total`

Counter · HTTP requests.

Counts completed requests to the metrics handler, including Prometheus scrapes and manual HTTP requests. The `code` label contains the response status; series for `200`, `500` and `503` are initialized at startup.

The counter increments after a response finishes, so a scrape does not include itself in the value it receives. Requests to `/getall` and `/getall.json` are outside this handler's instrumentation.

Metrics-handler HTTP errors per second:

```promql
sum by (instance, code) (rate(promhttp_metric_handler_requests_total{job="trace-monitor",code=~"5.."}[5m]))
```

### `promhttp_metric_handler_requests_in_flight`

Gauge · HTTP requests.

Counts overlapping requests to the metrics handler. The request reading this gauge is included, so a value of `1` during an otherwise idle scrape is expected.

Higher values show concurrent scrapers or requests that overlap in time. This gauge covers metrics requests only and carries no additional labels.

Concurrent metrics requests per Collector:

```promql
promhttp_metric_handler_requests_in_flight{job="trace-monitor"}
```

## Go runtime metrics

These describe the Collector's Go process, regardless of the languages sending traces. The definitions come from the client's [Go collector](https://github.com/prometheus/client_golang/blob/v1.14.0/prometheus/go_collector.go) and [runtime compatibility mapping](https://github.com/prometheus/client_golang/blob/v1.14.0/prometheus/go_collector_latest.go). Apart from `go_info` and the GC quantile series, they have no additional exporter labels. Library or Go upgrades can change the available runtime series.

### `go_info`

Gauge · constant 1.

Always exports `1` with a `version` label such as `go1.26.4`. Use the label to identify the Go toolchain behind a running binary; the Collector's application version is a separate value exposed by `/getall.json`.

### `go_goroutines`

Gauge · goroutines.

Reports the Go runtime's current goroutine count, including listeners, queue readers, HTTP handlers and runtime work. Compare trends under similar traffic to investigate accumulating work; this metric describes the Collector process, not the applications sending traces.

### `go_threads`

Gauge · OS threads.

Reports the number of OS threads recorded by the runtime's thread-creation profile. Go schedules goroutines onto OS threads, so this is a separate resource signal from `go_goroutines`; it is exported as a gauge.

### `go_gc_duration_seconds`

Summary · seconds.

Exports pause quantiles with `quantile` values `0`, `0.25`, `0.5`, `0.75` and `1` from the Go runtime's retained pause history. It also exports `go_gc_duration_seconds_sum` for cumulative pause seconds and `go_gc_duration_seconds_count` for cumulative GC cycles.

The quantiles have the runtime's history window, not a PromQL-selected time window, and cannot be meaningfully summed across instances. The sum and count can be used for a windowed mean when GC cycles occurred in that window.

Mean GC pause per cycle over five minutes, in seconds:

```promql
rate(go_gc_duration_seconds_sum{job="trace-monitor"}[5m])
/
rate(go_gc_duration_seconds_count{job="trace-monitor"}[5m])
```

### `go_memstats_alloc_bytes`

Gauge · bytes.

Reports allocated heap-object bytes, including objects that have become unreachable but have not yet been reclaimed. This is an alias of `go_memstats_heap_alloc_bytes` in the current exporter; use either series without adding them together.

### `go_memstats_alloc_bytes_total`

Counter · bytes.

Accumulates bytes allocated on the Go heap, including allocations that have since been freed. Its rate measures allocation pressure and helps explain GC activity; a large lifetime total does not establish a memory leak.

Heap allocation rate in MiB/s:

```promql
rate(go_memstats_alloc_bytes_total{job="trace-monitor"}[5m]) / 1024 / 1024
```

### `go_memstats_heap_alloc_bytes`

Gauge · bytes.

Reports the same heap-object allocation value as `go_memstats_alloc_bytes`. Compare it with `go_memstats_heap_inuse_bytes` to distinguish object bytes from the larger spans reserved for those objects.

### `go_memstats_heap_sys_bytes`

Gauge · bytes.

Includes both in-use and idle heap spans: `heap_sys_bytes = heap_inuse_bytes + heap_idle_bytes`. It can remain high after traffic falls because the runtime keeps heap space available for reuse.

### `go_memstats_heap_idle_bytes`

Gauge · bytes.

Measures idle heap spans, including memory already returned to the operating system. Subtract `go_memstats_heap_released_bytes` to estimate idle heap memory retained for reuse without returning it to the OS.

### `go_memstats_heap_inuse_bytes`

Gauge · bytes.

Measures whole heap spans used for objects, including unused space inside those spans. It can exceed `go_memstats_heap_alloc_bytes`; the difference helps inspect allocator slack and fragmentation.

### `go_memstats_heap_released_bytes`

Gauge · bytes.

Reports idle heap memory whose physical backing the runtime has released to the operating system. It is a subset of `go_memstats_heap_idle_bytes` and can decrease when the runtime reuses released space.

### `go_memstats_heap_objects`

Gauge · objects.

Counts heap objects that remain allocated, including unreachable objects awaiting garbage collection. Follow this alongside heap bytes and GC activity to distinguish many small objects from a smaller number of large allocations.

### `go_memstats_mallocs_total`

Counter · allocations.

Counts heap allocation events using the runtime's compatibility accounting, which includes tiny allocations. Its rate describes allocation frequency; compare with the byte-allocation rate to understand changes in allocation size.

### `go_memstats_frees_total`

Counter · frees.

Counts objects freed by the Go runtime under the same compatibility accounting as `go_memstats_mallocs_total`. It often rises around garbage collection; it is not a count of bytes returned to the operating system.

### `go_memstats_lookups_total`

Counter · lookups.

A legacy compatibility metric. The Go 1.17+ mapping in the pinned Prometheus client explicitly sets this value to zero, so it does not provide an activity or health signal for this service's supported Go versions.

### `go_memstats_stack_inuse_bytes`

Gauge · bytes.

Reports heap memory assigned to goroutine stacks. Growth can accompany an increase in goroutine count or deeper stacks; this memory is accounted separately from heap-object bytes.

### `go_memstats_stack_sys_bytes`

Gauge · bytes.

Includes goroutine stack memory and the OS-thread stack memory reported by the runtime. Compare it with `go_memstats_stack_inuse_bytes` when investigating stack-related memory usage.

### `go_memstats_mspan_inuse_bytes`

Gauge · bytes.

Measures memory used by the runtime's active `mspan` structures, which describe heap spans. It is allocator bookkeeping and contributes to runtime overhead rather than stored trace payload bytes.

### `go_memstats_mspan_sys_bytes`

Gauge · bytes.

Includes active `mspan` structures and reserved metadata space available for reuse. The difference from `go_memstats_mspan_inuse_bytes` shows retained capacity in this part of the allocator.

### `go_memstats_mcache_inuse_bytes`

Gauge · bytes.

Measures memory used by active `mcache` structures, the runtime's allocation caches. Changes reflect allocator infrastructure rather than a direct count of traces or application objects.

### `go_memstats_mcache_sys_bytes`

Gauge · bytes.

Includes active allocation-cache metadata and reserved space available for reuse. Compare it with `go_memstats_mcache_inuse_bytes` to inspect unused capacity in this metadata pool.

### `go_memstats_buck_hash_sys_bytes`

Gauge · bytes.

Reports memory assigned to profiling bucket hash tables. It helps account for profiling-related runtime overhead and is separate from the trace context and backtrace data stored by the application.

### `go_memstats_gc_sys_bytes`

Gauge · bytes.

Reports the legacy GC-metadata memory field. In this client's Go 1.17+ compatibility mapping it is populated from the runtime's other-metadata memory class; it describes bookkeeping bytes, not GC CPU usage or pause time.

### `go_memstats_other_sys_bytes`

Gauge · bytes.

Reports miscellaneous memory managed by the Go runtime outside the heap, stack and named metadata categories. Use it to help explain growth in total runtime memory that is not visible in the other component gauges.

### `go_memstats_sys_bytes`

Gauge · bytes.

Covers the runtime's heap, stacks and bookkeeping memory, including space retained for reuse or released within reserved mappings. For actual resident process memory, use `process_resident_memory_bytes` where the platform exposes it.

### `go_memstats_next_gc_bytes`

Gauge · bytes.

Reports the runtime's target heap size for the next garbage-collection cycle. It adapts to live memory and GC settings; it is not a hard memory limit for the Collector.

### `go_memstats_last_gc_time_seconds`

Gauge · Unix seconds.

Reports the Unix timestamp of the last completed GC. Subtract a nonzero value from `time()` to find the time since that collection; before a collection has been reported, a zero value is not a useful GC-age measurement.

## Process metrics

The default [process collector](https://github.com/prometheus/client_golang/blob/v1.14.0/prometheus/process_collector_other.go) reads these values from `/proc` on Linux. They describe the Collector process and have no additional exporter labels. Unsupported platforms or inaccessible process statistics can cause some or all of these series to be absent; absence is different from a zero value.

### `process_cpu_seconds_total`

Counter · CPU seconds.

Accumulates user and kernel CPU time consumed by the Collector process. `rate(...[5m])` expresses average CPU cores used: a value of `1` corresponds to one fully utilized core, and values above `1` are possible.

Average CPU cores used over five minutes:

```promql
rate(process_cpu_seconds_total{job="trace-monitor"}[5m])
```

### `process_resident_memory_bytes`

Gauge · bytes.

Reports the process's resident set size, covering physical memory residency beyond just Go heap objects. Use it to track process memory footprint; it can differ substantially from the Go runtime's reserved-memory gauges.

### `process_virtual_memory_bytes`

Gauge · bytes.

Reports the process's mapped or reserved virtual address-space size. Large reservations do not necessarily consume an equivalent amount of physical RAM, so interpret it alongside resident memory.

### `process_virtual_memory_max_bytes`

Gauge · bytes.

Reports the process's virtual address-space resource limit. An unlimited limit may appear as a very large value; this metric does not describe the machine's RAM or a container's memory limit.

### `process_open_fds`

Gauge · file descriptors.

Counts open files, sockets and other file descriptors held by the process. Compare it with `process_max_fds` to inspect descriptor headroom and investigate sustained growth.

Fraction of the file-descriptor limit in use:

```promql
process_open_fds{job="trace-monitor"} / process_max_fds{job="trace-monitor"}
```

### `process_max_fds`

Gauge · file descriptors.

Reports the process's open-file resource limit. It is a ceiling for descriptor usage, while `process_open_fds` provides the current usage; a change in this gauge can reflect a deployment or OS-limit change.

### `process_start_time_seconds`

Gauge · Unix seconds.

Reports when the current Collector process started. Use it to identify restarts and calculate process uptime; unlike a counter, the timestamp changes to a new epoch value when the process is replaced.

Process uptime in seconds:

```promql
time() - process_start_time_seconds{job="trace-monitor"}
```

## Coverage and limits

Trace/span durations, request latency, UDP byte volume, current queue depth, exact packet-loss counts and command-validation errors have no dedicated metrics in this version. Use the live trace endpoints and logs when investigating individual traces or rejected commands. In particular, the packet and trace command counters cannot establish application success rates.

Prometheus also creates target-health series such as `up` and `scrape_duration_seconds` while scraping. These are generated by Prometheus and are not served by the Collector's `/console/metrics` endpoint.

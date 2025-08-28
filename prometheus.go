package main

import (
	"os"
	"strings"
	"trace-monitor-collector/config"
	"trace-monitor-collector/traceCollection"

	"github.com/prometheus/client_golang/prometheus"
)

type metricsStruct struct {
	cfg                 *config.Config
	TotalTraceSet       *prometheus.Desc
	TotalSpanSet        *prometheus.Desc
	TotalAllSpanClose   *prometheus.Desc
	TotalTraceDelete    *prometheus.Desc
	TotalPackagesCaught *prometheus.Desc
	TotalPackagesParse  *prometheus.Desc
	CountActivePid      *prometheus.Desc
	TotalChannelReset   *prometheus.Desc
}

func NewExporter(cfg *config.Config) *metricsStruct {
	return &metricsStruct{
		cfg: cfg,
		TotalTraceSet: prometheus.NewDesc("trace_monitor_total_trace_set",
			"Total traces processed by trace Monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalSpanSet: prometheus.NewDesc("trace_monitor_total_span_set",
			"Total number of spans set in trace monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalAllSpanClose: prometheus.NewDesc("trace_monitor_total_all_span_close",
			"Total count of closed spans in trace Monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalTraceDelete: prometheus.NewDesc("trace_monitor_total_trace_delete",
			"Total deleted traces count in trace Monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalPackagesCaught: prometheus.NewDesc("trace_monitor_total_packages_caught",
			"Total caught packages monitored by trace monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalPackagesParse: prometheus.NewDesc("trace_monitor_total_packages_parse",
			"Total caught packages monitored by trace monitor",
			[]string{"node", "app", "env"},
			nil,
		),
		CountActivePid: prometheus.NewDesc("trace_monitor_count_active_pid",
			"Number of active PIDs in the trace monitoring",
			[]string{"node", "app", "env"},
			nil,
		),
		TotalChannelReset: prometheus.NewDesc("trace_monitor_total_channel_reset",
			"Total resets of the package channel",
			[]string{"node", "app", "env"},
			nil,
		),
	}
}

func (collector *metricsStruct) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(collector, ch)
}

func (collector *metricsStruct) Collect(ch chan<- prometheus.Metric) {
	hostname, _ := os.Hostname()
	node := strings.Split(hostname, ".")[0]
	app := collector.cfg.AppName
	env := collector.cfg.Env

	m1 := prometheus.MustNewConstMetric(collector.TotalTraceSet, prometheus.CounterValue, float64(traceCollection.TotalTraceSet.Count()), node, app, env)
	ch <- m1
	m2 := prometheus.MustNewConstMetric(collector.TotalSpanSet, prometheus.CounterValue, float64(traceCollection.TotalSpanSet.Count()), node, app, env)
	ch <- m2
	m3 := prometheus.MustNewConstMetric(collector.TotalAllSpanClose, prometheus.CounterValue, float64(traceCollection.TotalAllSpanClose.Count()), node, app, env)
	ch <- m3
	m4 := prometheus.MustNewConstMetric(collector.TotalTraceDelete, prometheus.CounterValue, float64(traceCollection.TotalTraceDelete.Count()), node, app, env)
	ch <- m4
	m5 := prometheus.MustNewConstMetric(collector.TotalPackagesCaught, prometheus.CounterValue, float64(totalPackagesCaught.Count()), node, app, env)
	ch <- m5
	m6 := prometheus.MustNewConstMetric(collector.TotalPackagesParse, prometheus.CounterValue, float64(totalPackagesParse.Count()), node, app, env)
	ch <- m6
	m7 := prometheus.MustNewConstMetric(collector.CountActivePid, prometheus.GaugeValue, float64(traceCollection.CountActivePid.Count()), node, app, env)
	ch <- m7
	m8 := prometheus.MustNewConstMetric(collector.TotalChannelReset, prometheus.CounterValue, float64(totalChannelReset.Count()), node, app, env)
	ch <- m8
}

func handlePrometheus(cfg *config.Config) {
	foo := NewExporter(cfg)
	prometheus.MustRegister(foo)
}

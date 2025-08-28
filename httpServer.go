package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"trace-monitor-collector/config"
	"trace-monitor-collector/traceCollection"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed json-viewer.tmpl
var tmpl embed.FS

type dataStruct struct {
	Trace   []byte
	Span    []byte
	Context []byte
	Tags    []byte
	TraceId string
	SentAt  time.Time
}

func handleHttp(cfg *config.Config) {
	defer recoverRoutineHandleHttp(cfg)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		routeHTTP(w, r, cfg)
	})
	http.Handle("/console/metrics", promhttp.Handler())
	if cfg.IsVerboseByLevel("v") {
		log.Println("HTTP server started on", cfg.HttpAddr)
	}
	http.ListenAndServe(cfg.HttpAddr, nil)
}

func recoverRoutineHandleHttp(cfg *config.Config) {
	if r := recover(); r != nil {
		log.Println("Handle HTTP server error: ", r)
		go handleHttp(cfg)
	}
}

func routeHTTP(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.URL.Path == "/console/metrics" {
		promhttp.Handler().ServeHTTP(w, r)
	} else if r.URL.Path == "/getall.json" {
		jsonBytes, err := buildJsonBytesAll(cfg)
		if err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}
		w.Write(jsonBytes)
	} else if r.URL.Path == "/getall" {
		jsonBytes, err := buildJsonBytesAll(cfg)
		if err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}
		t, err := template.ParseFS(tmpl, "json-viewer.tmpl")
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(w, string(jsonBytes))
		if err != nil {
			log.Printf("Error execute json to template: %v", err)
		}
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}
}

func buildJsonBytesAll(cfg *config.Config) ([]byte, error) {
	jsonData := map[string]map[string]map[string]interface{}{
		"stats": {
			"_": {
				"serverTime": time.Now().String(),
				"appVersion": AppVersion,
			},
			"totalCounts": {
				"traceSet":          traceCollection.TotalTraceSet.Count(),
				"spanSet":           traceCollection.TotalSpanSet.Count(),
				"allSpanClose":      traceCollection.TotalAllSpanClose.Count(),
				"traceDelete":       traceCollection.TotalTraceDelete.Count(),
				"packagesCaught":    totalPackagesCaught.Count(),
				"packagesParse":     totalPackagesParse.Count(),
				"totalChannelReset": totalChannelReset.Count(),
			},
			"gauge": {
				"countActivePid": traceCollection.CountActivePid.Count(),
			},
		},
		"trace": {},
	}
	dataCollection := traceCollection.GetAllTrace()
	for pid, valueByte := range dataCollection {
		var valueData = dataStruct{}
		json.Unmarshal(valueByte, &valueData)
		duration := time.Since(valueData.SentAt)
		var trace map[string]interface{}
		json.Unmarshal(valueData.Trace, &trace)
		var trace_context map[string]interface{}
		var trace_tags map[string]interface{}
		if trace != nil {
			trace = unpackData(trace)
			trace_context, _ = trace["context"].(map[string]interface{})
			trace_tags, _ = trace["tags"].(map[string]interface{})
			delete(trace, "context")
			delete(trace, "tags")
		}
		var span map[string]interface{}
		json.Unmarshal(valueData.Span, &span)
		if span != nil {
			span = unpackData(span)
		}
		var context map[string]interface{}
		json.Unmarshal(valueData.Context, &context)

		if context == nil {
			context = trace_context
		} else {
			context = unpackData(context)
		}
		var tags map[string]interface{}
		json.Unmarshal(valueData.Tags, &tags)
		if tags == nil {
			tags = trace_tags
		} else {
			tags = unpackData(tags)
		}
		pidInfo := map[string]interface{}{
			"sentAt":      valueData.SentAt,
			"pid":         pid,
			"traceId":     valueData.TraceId,
			"elapsedTime": duration.String(),
			"trace":       trace,
			"span":        span,
			"context":     context,
			"tags":        tags,
		}
		jsonData["trace"][pid] = pidInfo
	}
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return []byte{}, fmt.Errorf("skip delete trace command. %v", err)
	}
	return jsonBytes, nil
}

func unpackData(data map[string]interface{}) map[string]interface{} {
	if _, ok := data["data"]; ok {
		return data["data"].(map[string]interface{})
	} else {
		return data
	}
}

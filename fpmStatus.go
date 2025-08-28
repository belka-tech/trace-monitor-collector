package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"trace-monitor-collector/config"
	"trace-monitor-collector/traceCollection"
)

func loadFpmStatus(cfg *config.Config) (map[string]interface{}, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   cfg.HttpClientTimeout * time.Second,
		Transport: transport,
	}
	var fpmStatus map[string]interface{}
	resp, err := client.Get(cfg.FpmStatusURL)
	if err != nil {
		return fpmStatus, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fpmStatus, err
	}
	if cfg.IsVerboseByLevel("vvv") {
		log.Println("FPM Status respons body", string(body))
	}
	jsonErr := json.Unmarshal(body, &fpmStatus)
	if jsonErr != nil {
		return fpmStatus, jsonErr
	}

	return fpmStatus, nil
}

func buildPidMap(fpmStatus map[string]interface{}) map[string]map[string]interface{} {
	var fpmStatusPidMap = make(map[string]map[string]interface{})
	for _, process := range fpmStatus["processes"].([]interface{}) {
		pidFloat := process.(map[string]interface{})["pid"].(float64)
		pid := strconv.FormatFloat(pidFloat, 'f', -1, 64)
		fpmStatusPidMap[pid] = process.(map[string]interface{})
	}
	return fpmStatusPidMap
}

func handleFpmStatus(cfg *config.Config) {
	defer recoverRoutineHandleFpmStatus(cfg)

	ticker := time.NewTicker(cfg.LoadFpmStatusTimeout * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		fpmStatus, err := loadFpmStatus(cfg)
		if err != nil {
			if cfg.IsVerboseByLevel("v") {
				log.Println("load FPM status error.", err)
			}
			continue
		}
		fpmStatusPidMap := buildPidMap(fpmStatus)

		if cfg.IsVerboseByLevel("vvv") {
			log.Println("Build pid map from fpm status", fpmStatusPidMap)
		}
		traceCollection.CheckingForHung(cfg, fpmStatusPidMap)
	}
}

func recoverRoutineHandleFpmStatus(cfg *config.Config) {
	if r := recover(); r != nil {
		log.Println("Handle FPM Status error: ", r)
		go handleFpmStatus(cfg)
	}
}

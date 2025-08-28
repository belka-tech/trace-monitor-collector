package traceCollection

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"trace-monitor-collector/config"
	"trace-monitor-collector/counter"
)

type dataStruct struct {
	TraceId string
	SentAt  time.Time

	Trace   []byte
	Span    []byte
	Context []byte // Поле используется в httpServer перед выпиливанием проверить там
	Tags    []byte // Поле используется в httpServer перед выпиливанием проверить там
}

type ChronologicalError struct {
	Err error
}

func (e *ChronologicalError) Error() string {
	return e.Err.Error()
}

var (
	dataCollection    sync.Map
	TotalTraceSet     counter.CounterStruct
	TotalSpanSet      counter.CounterStruct
	TotalTraceDelete  counter.CounterStruct
	TotalAllSpanClose counter.CounterStruct
	CountActivePid    counter.CounterStruct
)

func isChronologicalCorrect(traceData *dataStruct, newTime time.Time) (bool, error) {
	if !traceData.SentAt.Before(newTime) {
		return false, fmt.Errorf("message history is broken")
	}
	return true, nil
}

func isTraceIdIdentical(traceData *dataStruct, traceId string) bool {
	return traceData.TraceId == traceId
}

func createTraceData(pid string, traceId string) *dataStruct {
	newTraceData := new(dataStruct)
	newTraceData.TraceId = traceId

	dataCollection.Store(pid, &newTraceData)
	CountActivePid.Increment()

	return newTraceData
}

func deleteTraceData(pid string) {
	dataCollection.Delete(pid)
	CountActivePid.Decrement()
}

func InitTrace(cfg *config.Config, pid string, traceId string, sentAt time.Time, data []byte) error {
	TotalTraceSet.Increment()
	var traceData *dataStruct
	if value, isExist := dataCollection.Load(pid); isExist {
		traceData = *value.(**dataStruct)
		if isTraceIdOk := isTraceIdIdentical(traceData, traceId); !isTraceIdOk {
			if isChronologicalOk, err := isChronologicalCorrect(traceData, sentAt); !isChronologicalOk {
				return fmt.Errorf("skip set trace command. %v", err)
			}
			if cfg.IsVerboseByLevel("v") {
				log.Println("_warn: _", pid, "new trace without deleting `SetTrace`", traceId)
			}
			deleteTraceData(pid)
			traceData = createTraceData(pid, traceId)
			traceData.SentAt = sentAt
		}
	} else {
		traceData = createTraceData(pid, traceId)
		traceData.SentAt = sentAt
	}
	traceData.Trace = data
	dataCollection.Store(pid, &traceData)

	return nil
}

func SetTraceCurrentSpan(cfg *config.Config, pid string, traceId string, sentAt time.Time, data []byte) error {
	TotalSpanSet.Increment()
	var traceData *dataStruct
	if value, isExist := dataCollection.Load(pid); isExist {
		traceData = *value.(**dataStruct)
		if isChronologicalOk, err := isChronologicalCorrect(traceData, sentAt); !isChronologicalOk {
			return fmt.Errorf("skip set span command. %v", err)
		}
		if isTraceIdOk := isTraceIdIdentical(traceData, traceId); !isTraceIdOk {
			if cfg.IsVerboseByLevel("v") {
				log.Println("_warn: _", pid, "new trace without deleting `SetSpan`", traceId)
			}
			deleteTraceData(pid)
			traceData = createTraceData(pid, traceId)
		}
	} else {
		traceData = createTraceData(pid, traceId)
	}

	traceData.SentAt = sentAt
	traceData.Span = data
	dataCollection.Store(pid, &traceData)

	return nil
}

func DeleteSpan(cfg *config.Config, pid string, traceId string, sentAt time.Time) error {
	TotalAllSpanClose.Increment()
	var traceData *dataStruct
	if value, isExist := dataCollection.Load(pid); isExist {
		traceData = *value.(**dataStruct)
		if isChronologicalOk, err := isChronologicalCorrect(traceData, sentAt); !isChronologicalOk {
			return fmt.Errorf("skip delete span command. %v", err)
		}
		if isTraceIdOk := isTraceIdIdentical(traceData, traceId); isTraceIdOk {
			traceData.Span = nil
			dataCollection.Store(pid, &traceData)
		} else {
			if cfg.IsVerboseByLevel("v") {
				log.Println("_warn: _", pid, "new trace without deleting `DeleteSpan`", traceId)
			}
			deleteTraceData(pid)
		}
	}

	return nil
}

func DeleteTrace(cfg *config.Config, pid string, traceId string, sentAt time.Time) error {
	TotalTraceDelete.Increment()
	var traceData *dataStruct
	if value, isExist := dataCollection.Load(pid); isExist {
		traceData = *value.(**dataStruct)
		if isChronologicalOk, err := isChronologicalCorrect(traceData, sentAt); !isChronologicalOk {
			return fmt.Errorf("skip delete trace command. %v", err)
		}
		if isTraceIdOk := isTraceIdIdentical(traceData, traceId); !isTraceIdOk {
			if cfg.IsVerboseByLevel("v") {
				log.Println("_warn: _", pid, "new trace without deleting `DeleteTrace`", traceId)
			}
		}
		deleteTraceData(pid)
	}

	return nil
}

func GetAllTrace() map[string][]byte {
	var localTraceCollection = make(map[string][]byte)
	dataCollection.Range(func(pid, value interface{}) bool {
		jsonBytes, _ := json.Marshal(**value.(**dataStruct))
		localTraceCollection[pid.(string)] = jsonBytes
		return true
	})
	return localTraceCollection
}

func CheckingForHung(cfg *config.Config, fpmStatusPIDmap map[string]map[string]interface{}) {
	dataCollection.Range(func(Pid, value interface{}) bool {
		valueData := **value.(**dataStruct)
		localPid := Pid.(string)
		if time.Since(valueData.SentAt) < (cfg.StuckProcessDuration * time.Second) {
			return true
		}
		if cfg.IsVerboseByLevel("v") {
			log.Printf("CheckingForHung - %v", localPid)
		}
		pidInfo, isExist := fpmStatusPIDmap[localPid]
		if !isExist {
			if cfg.IsVerboseByLevel("v") {
				log.Println("Process PID missing:", localPid, valueData.TraceId)
			}
			deleteTraceData(localPid)
		} else if pidInfo["state"] == "Idle" {
			if cfg.IsVerboseByLevel("v") {
				log.Println("delete by Idle")
			}
			deleteTraceData(localPid)
		}
		return true
	})
}

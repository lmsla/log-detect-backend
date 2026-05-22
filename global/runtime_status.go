package global

import "sync"

type MySQLRuntimeStatus struct {
	Configured bool   `json:"configured"`
	Connected  bool   `json:"connected"`
	Host       string `json:"host,omitempty"`
	Port       string `json:"port,omitempty"`
	Database   string `json:"database,omitempty"`
}

type TimescaleRuntimeStatus struct {
	Configured       bool     `json:"configured"`
	Connected        bool     `json:"connected"`
	Host             string   `json:"host,omitempty"`
	Port             string   `json:"port,omitempty"`
	Error            string   `json:"error,omitempty"`
	DegradedFeatures []string `json:"degraded_features,omitempty"`
	LastCheckAt      string   `json:"last_check_at,omitempty"`
}

type ESMonitoringRuntimeStatus struct {
	Enabled          bool   `json:"enabled"`
	SchedulerStarted bool   `json:"scheduler_started"`
	Reason           string `json:"reason,omitempty"`
}

type SystemRuntimeStatus struct {
	MySQL        MySQLRuntimeStatus        `json:"mysql"`
	TimescaleDB  TimescaleRuntimeStatus    `json:"timescaledb"`
	ESMonitoring ESMonitoringRuntimeStatus `json:"es_monitoring"`
}

var (
	runtimeStatus   SystemRuntimeStatus
	runtimeStatusMu sync.RWMutex
)

func GetRuntimeStatus() SystemRuntimeStatus {
	runtimeStatusMu.RLock()
	defer runtimeStatusMu.RUnlock()
	return runtimeStatus
}

func SetMySQLRuntimeStatus(status MySQLRuntimeStatus) {
	runtimeStatusMu.Lock()
	defer runtimeStatusMu.Unlock()
	runtimeStatus.MySQL = status
}

func SetTimescaleRuntimeStatus(status TimescaleRuntimeStatus) {
	runtimeStatusMu.Lock()
	defer runtimeStatusMu.Unlock()
	status.DegradedFeatures = append([]string(nil), status.DegradedFeatures...)
	runtimeStatus.TimescaleDB = status
}

func SetESMonitoringRuntimeStatus(status ESMonitoringRuntimeStatus) {
	runtimeStatusMu.Lock()
	defer runtimeStatusMu.Unlock()
	runtimeStatus.ESMonitoring = status
}

package services

import (
	"fmt"
	"log-detect/global"
)

func IsTimescaleReady() bool {
	status := global.GetRuntimeStatus().TimescaleDB
	return status.Configured &&
		status.Connected &&
		global.TimescaleDB != nil
}

func TimescaleDependentFeatures() []string {
	features := []string{"history"}
	if global.EnvConfig != nil && global.EnvConfig.Features.Dashboard {
		features = append(features, "dashboard")
	}
	if global.EnvConfig != nil && global.EnvConfig.Features.ESMonitoring {
		features = append(features, "es_monitoring")
	}
	return features
}

func TimescaleUnavailableMessage() string {
	status := global.GetRuntimeStatus().TimescaleDB
	if status.Error != "" {
		return fmt.Sprintf("TimescaleDB unavailable for this run: %s", status.Error)
	}
	if !status.Configured {
		return "TimescaleDB feature is disabled in setting.yml"
	}
	return "TimescaleDB unavailable for this run; history and Timescale-dependent features are disabled"
}

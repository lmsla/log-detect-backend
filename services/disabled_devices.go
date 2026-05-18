package services

import (
	"log-detect/global"
	"log-detect/structs"
)

func IsDeviceDisabled(deviceGroup, deviceName string) bool {
	cfg := global.GetYMLConfig()
	return isDeviceDisabledInConfig(cfg, deviceGroup, deviceName)
}

func isDeviceDisabledInConfig(cfg *structs.YMLConfig, deviceGroup, deviceName string) bool {
	if cfg == nil || deviceGroup == "" || deviceName == "" {
		return false
	}

	for _, group := range cfg.DisabledDevices {
		if group.DeviceGroup != deviceGroup {
			continue
		}
		for _, device := range group.Devices {
			if device.Name == deviceName {
				return true
			}
		}
	}

	return false
}

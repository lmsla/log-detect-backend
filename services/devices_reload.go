package services

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"log-detect/global"
	"log-detect/structs"
	"log-detect/utils"
	"os"
	"time"
)

func StartDevicesYMLReload() {
	if global.EnvConfig == nil || global.EnvConfig.ConfigSource != "yml" {
		return
	}

	if !global.EnvConfig.YMLReload.DevicesEnabled {
		log.Println("devices.yml hot reload is disabled")
		return
	}

	interval := 30 * time.Second
	if global.EnvConfig.YMLReload.DevicesInterval != "" {
		parsed, err := time.ParseDuration(global.EnvConfig.YMLReload.DevicesInterval)
		if err != nil {
			log.Printf("Invalid yml_reload.devices_interval '%s', fallback to 30s",
				global.EnvConfig.YMLReload.DevicesInterval)
		} else if parsed > 0 {
			interval = parsed
		}
	}

	lastChecksum, err := readDevicesFileChecksum()
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("devices.yml hot reload enabled (interval=%s), waiting for devices.yml to appear", interval)
		} else {
			log.Printf("devices.yml hot reload enabled (interval=%s), but initial checksum read failed: %v", interval, err)
		}
	} else {
		log.Printf("devices.yml hot reload enabled (interval=%s)", interval)
	}

	go runDevicesYMLReloadLoop(interval, lastChecksum)
}

func runDevicesYMLReloadLoop(interval time.Duration, lastChecksum string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		devicesConfig, data, err := utils.ReadDevicesFileConfig()
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("devices.yml hot reload skipped due to read error: %v", err)
			}
			continue
		}

		checksum := checksumBytes(data)
		if checksum == lastChecksum {
			continue
		}

		nextCfg := mergeDevicesConfig(global.GetYMLConfig(), devicesConfig)
		if err := SyncDevicesConfigToDB(nextCfg); err != nil {
			log.Printf("devices.yml hot reload sync failed: %v", err)
			continue
		}

		global.SetYMLConfig(nextCfg)
		lastChecksum = checksum
		log.Printf("devices.yml hot reload applied: %d device groups, %d disabled groups",
			len(nextCfg.Devices), len(nextCfg.DisabledDevices))
	}
}

func readDevicesFileChecksum() (string, error) {
	_, data, err := utils.ReadDevicesFileConfig()
	if err != nil {
		return "", err
	}
	return checksumBytes(data), nil
}

func checksumBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func mergeDevicesConfig(base, devicesCfg *structs.YMLConfig) *structs.YMLConfig {
	next := &structs.YMLConfig{}
	if base != nil {
		next.ESConnections = append(next.ESConnections, base.ESConnections...)
		next.Targets = append(next.Targets, base.Targets...)
	}
	if devicesCfg != nil {
		next.Devices = cloneDeviceGroups(devicesCfg.Devices)
		next.DisabledDevices = cloneDisabledDeviceGroups(devicesCfg.DisabledDevices)
	}
	return next
}

func cloneDeviceGroups(groups []structs.YMLDeviceGroup) []structs.YMLDeviceGroup {
	cloned := make([]structs.YMLDeviceGroup, 0, len(groups))
	for _, group := range groups {
		nextGroup := structs.YMLDeviceGroup{
			DeviceGroup: group.DeviceGroup,
			Names:       append([]structs.YMLDeviceItem(nil), group.Names...),
		}
		cloned = append(cloned, nextGroup)
	}
	return cloned
}

func cloneDisabledDeviceGroups(groups []structs.YMLDisabledDeviceGroup) []structs.YMLDisabledDeviceGroup {
	cloned := make([]structs.YMLDisabledDeviceGroup, 0, len(groups))
	for _, group := range groups {
		nextGroup := structs.YMLDisabledDeviceGroup{
			DeviceGroup: group.DeviceGroup,
			Devices:     append([]structs.YMLDisabledDevice(nil), group.Devices...),
		}
		cloned = append(cloned, nextGroup)
	}
	return cloned
}

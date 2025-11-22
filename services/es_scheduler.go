package services

import (
	"fmt"
	"log-detect/entities"
	"log-detect/log"
	"sync"
	"time"
)

// ESMonitorScheduler 管理所有 ES 監控器的定時任務
type ESMonitorScheduler struct {
	monitors map[int]*ESMonitorJob // monitor_id -> job
	mutex    sync.RWMutex
}

// ESMonitorJob 代表單個監控器的定時任務
type ESMonitorJob struct {
	Monitor  entities.ElasticsearchMonitor
	Ticker   *time.Ticker
	StopChan chan bool
}

var (
	GlobalESScheduler *ESMonitorScheduler
	once              sync.Once
)

// InitESScheduler 初始化全域排程器（單例模式）
func InitESScheduler() {
	once.Do(func() {
		GlobalESScheduler = &ESMonitorScheduler{
			monitors: make(map[int]*ESMonitorJob),
		}
		log.Logrecord_no_rotate("INFO", "ES Monitor Scheduler initialized")
	})
}

// LoadAllMonitors 從資料庫載入所有啟用的監控器並啟動排程
func (s *ESMonitorScheduler) LoadAllMonitors() error {
	log.Logrecord_no_rotate("INFO", "Loading all enabled ES monitors...")

	// 從資料庫載入所有監控器
	result := GetAllESMonitors()

	if !result.Success {
		return fmt.Errorf("failed to load monitors: %s", result.Msg)
	}

	monitors, ok := result.Body.([]entities.ElasticsearchMonitor)
	if !ok {
		return fmt.Errorf("invalid monitors data type")
	}

	// 啟動所有已啟用的監控器
	count := 0
	for _, monitor := range monitors {
		if monitor.EnableMonitor {
			if err := s.StartMonitor(monitor); err != nil {
				log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to start monitor %d: %s", monitor.ID, err.Error()))
			} else {
				count++
			}
		}
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Successfully started %d ES monitors", count))
	return nil
}

// StartMonitor 啟動單個監控器的定時任務
func (s *ESMonitorScheduler) StartMonitor(monitor entities.ElasticsearchMonitor) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 檢查是否已存在
	if _, exists := s.monitors[monitor.ID]; exists {
		return fmt.Errorf("monitor %d already running", monitor.ID)
	}

	// 建立定時器
	interval := time.Duration(monitor.Interval) * time.Second
	ticker := time.NewTicker(interval)
	stopChan := make(chan bool)

	job := &ESMonitorJob{
		Monitor:  monitor,
		Ticker:   ticker,
		StopChan: stopChan,
	}

	s.monitors[monitor.ID] = job

	// 立即執行一次（不等待第一個 tick）
	go func() {
		esService := NewESMonitorService()
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Initial monitoring check for ES: %s (ID: %d)", monitor.Name, monitor.ID))
		esService.MonitorESCluster(monitor)
	}()

	// 啟動定時任務
	go func() {
		esService := NewESMonitorService()
		for {
			select {
			case <-ticker.C:
				log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Scheduled monitoring check for ES: %s (ID: %d)", monitor.Name, monitor.ID))
				esService.MonitorESCluster(monitor)
			case <-stopChan:
				ticker.Stop()
				log.Logrecord_no_rotate("INFO", fmt.Sprintf("Stopped monitoring ES: %s (ID: %d)", monitor.Name, monitor.ID))
				return
			}
		}
	}()

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Started monitoring ES: %s (ID: %d) with interval: %d seconds", monitor.Name, monitor.ID, monitor.Interval))
	return nil
}

// StopMonitor 停止單個監控器的定時任務
func (s *ESMonitorScheduler) StopMonitor(monitorID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, exists := s.monitors[monitorID]
	if !exists {
		return fmt.Errorf("monitor %d not found in scheduler", monitorID)
	}

	// 發送停止信號
	close(job.StopChan)

	// 從 map 中移除
	delete(s.monitors, monitorID)

	return nil
}

// RestartMonitor 重啟單個監控器（用於更新配置後）
func (s *ESMonitorScheduler) RestartMonitor(monitor entities.ElasticsearchMonitor) error {
	// 先停止（如果存在）
	if err := s.StopMonitor(monitor.ID); err != nil {
		// 如果不存在就忽略錯誤
		log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d was not running, starting fresh", monitor.ID))
	}

	// 只有啟用狀態才重新啟動
	if monitor.EnableMonitor {
		return s.StartMonitor(monitor)
	}

	return nil
}

// GetRunningMonitors 取得目前正在運行的監控器清單
func (s *ESMonitorScheduler) GetRunningMonitors() []int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	monitorIDs := make([]int, 0, len(s.monitors))
	for id := range s.monitors {
		monitorIDs = append(monitorIDs, id)
	}

	return monitorIDs
}

// GetMonitorStatus 取得特定監控器的運行狀態
func (s *ESMonitorScheduler) GetMonitorStatus(monitorID int) (bool, *entities.ElasticsearchMonitor) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	job, exists := s.monitors[monitorID]
	if !exists {
		return false, nil
	}

	return true, &job.Monitor
}

// StopAll 停止所有監控器（用於應用關閉）
func (s *ESMonitorScheduler) StopAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Logrecord_no_rotate("INFO", "Stopping all ES monitors...")

	for id, job := range s.monitors {
		close(job.StopChan)
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Stopped ES monitor: %d", id))
	}

	// 清空 map
	s.monitors = make(map[int]*ESMonitorJob)

	log.Logrecord_no_rotate("INFO", "All ES monitors stopped")
}

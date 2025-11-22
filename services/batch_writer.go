package services

import (
	"database/sql"
	"fmt"
	"log-detect/entities"
	"log-detect/log"
	"sync"
	"time"
)

// BatchWriter 批量寫入服務
type BatchWriter struct {
	db            *sql.DB
	batch         []entities.History
	esBatch       []entities.ESMetric
	batchSize     int
	flushInterval time.Duration
	mutex         sync.Mutex
	ticker        *time.Ticker
	stopChan      chan struct{}
	stmt          *sql.Stmt
	esStmt        *sql.Stmt
}

// NewBatchWriter 創建批量寫入服務
func NewBatchWriter(db *sql.DB, batchSize int, flushInterval time.Duration) *BatchWriter {
	bw := &BatchWriter{
		db:            db,
		batch:         make([]entities.History, 0, batchSize),
		esBatch:       make([]entities.ESMetric, 0, batchSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		ticker:        time.NewTicker(flushInterval),
		stopChan:      make(chan struct{}),
	}

	// 預編譯 device_metrics SQL 語句
	var err error
	bw.stmt, err = db.Prepare(`
		INSERT INTO device_metrics
		(time, device_id, device_group, logname, status, lost, lost_num,
		 date, hour_time, date_time, timestamp_unix, period, unit,
		 target_id, index_id, response_time, data_count, error_msg, error_code, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to prepare batch insert statement: %s", err.Error()))
	}

	// 預編譯 es_metrics SQL 語句
	bw.esStmt, err = db.Prepare(`
		INSERT INTO es_metrics
		(time, monitor_id, status, cluster_name, cluster_status, response_time,
		 cpu_usage, memory_usage, disk_usage, node_count, data_node_count,
		 query_latency, indexing_rate, search_rate, total_indices, total_documents,
		 total_size_bytes, active_shards, relocating_shards, unassigned_shards,
		 error_message, warning_message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`)
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to prepare ES batch insert statement: %s", err.Error()))
		log.Logrecord_no_rotate("WARN", "ES metrics batch writing will be disabled. Please check if es_metrics table exists and has all required columns.")
	} else {
		log.Logrecord_no_rotate("INFO", "✅ ES metrics prepared statement initialized successfully")
	}

	// 啟動定時刷新協程
	go bw.startFlushRoutine()

	return bw
}

// startFlushRoutine 定時刷新數據
func (bw *BatchWriter) startFlushRoutine() {
	for {
		select {
		case <-bw.ticker.C:
			bw.flushBatch()
		case <-bw.stopChan:
			bw.flushBatch()
			if bw.stmt != nil {
				bw.stmt.Close()
			}
			return
		}
	}
}

// AddHistory 添加歷史記錄到批次（支援多種類型）
func (bw *BatchWriter) AddHistory(history any) error {
	bw.mutex.Lock()
	defer bw.mutex.Unlock()

	// 判斷類型並添加到對應批次
	switch v := history.(type) {
	case entities.History:
		bw.batch = append(bw.batch, v)
		if len(bw.batch) >= bw.batchSize {
			go bw.flushDeviceMetrics()
		}
	case entities.ESMetric:
		bw.esBatch = append(bw.esBatch, v)
		if len(bw.esBatch) >= bw.batchSize {
			go bw.flushESMetrics()
		}
	default:
		return fmt.Errorf("unsupported history type: %T", history)
	}

	return nil
}

// flushBatch 刷新所有批次數據到數據庫
func (bw *BatchWriter) flushBatch() {
	bw.mutex.Lock()
	defer bw.mutex.Unlock()

	bw.flushDeviceMetrics()
	bw.flushESMetrics()
}

// flushDeviceMetrics 刷新設備監控批次
func (bw *BatchWriter) flushDeviceMetrics() {
	if len(bw.batch) == 0 {
		return
	}

	tx, err := bw.db.Begin()
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to begin transaction: %s", err.Error()))
		return
	}
	defer tx.Rollback()

	txStmt := tx.Stmt(bw.stmt)

	successCount := 0
	for _, h := range bw.batch {
		t := time.Unix(h.Timestamp, 0)
		lost := h.Lost == "true"
		metadata := h.Metadata
		if metadata == "" {
			metadata = "{}"
		}

		_, err := txStmt.Exec(
			t, h.Name, h.DeviceGroup, h.Logname,
			h.Status, lost, h.LostNum,
			h.Date, h.Time, h.DateTime, h.Timestamp, h.Period, h.Unit,
			h.TargetID, h.IndexID, h.ResponseTime, h.DataCount,
			h.ErrorMsg, h.ErrorCode, metadata,
		)

		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to insert history record: %s", err.Error()))
			continue
		}
		successCount++
	}

	if err := tx.Commit(); err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to commit batch: %s", err.Error()))
		return
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("✅ Successfully flushed %d/%d history records to TimescaleDB", successCount, len(bw.batch)))

	bw.batch = bw.batch[:0]
}

// flushESMetrics 刷新 ES 監控批次
func (bw *BatchWriter) flushESMetrics() {
	if len(bw.esBatch) == 0 {
		return
	}

	// 檢查 esStmt 是否已初始化
	if bw.esStmt == nil {
		log.Logrecord_no_rotate("ERROR", "ES prepared statement is nil, skipping flush")
		bw.esBatch = bw.esBatch[:0] // 清空批次避免重複嘗試
		return
	}

	tx, err := bw.db.Begin()
	if err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to begin ES transaction: %s", err.Error()))
		return
	}
	defer tx.Rollback()

	txStmt := tx.Stmt(bw.esStmt)

	successCount := 0
	for _, m := range bw.esBatch {
		_, err := txStmt.Exec(
			m.Time, m.MonitorID, m.Status, m.ClusterName, m.ClusterStatus, m.ResponseTime,
			m.CPUUsage, m.MemoryUsage, m.DiskUsage, m.NodeCount, m.DataNodeCount,
			m.QueryLatency, m.IndexingRate, m.SearchRate, m.TotalIndices, m.TotalDocuments,
			m.TotalSizeBytes, m.ActiveShards, m.RelocatingShards, m.UnassignedShards,
			m.ErrorMessage, m.WarningMessage, m.Metadata,
		)

		if err != nil {
			log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to insert ES metric: %s", err.Error()))
			continue
		}
		successCount++
	}

	if err := tx.Commit(); err != nil {
		log.Logrecord_no_rotate("ERROR", fmt.Sprintf("Failed to commit ES batch: %s", err.Error()))
		return
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("✅ Successfully flushed %d/%d ES metrics to TimescaleDB", successCount, len(bw.esBatch)))

	bw.esBatch = bw.esBatch[:0]
}

// Stop 停止批量寫入服務
func (bw *BatchWriter) Stop() {
	bw.ticker.Stop()
	close(bw.stopChan)
}

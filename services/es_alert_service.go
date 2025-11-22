package services

import (
	"database/sql"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/models"
	"time"

	"github.com/lib/pq"
)

// ESAlertService ES 告警管理服務
type ESAlertService struct{}

// NewESAlertService 創建告警服務實例
func NewESAlertService() *ESAlertService {
	return &ESAlertService{}
}

// GetAlerts 獲取告警列表（支援過濾與分頁）
func (s *ESAlertService) GetAlerts(params models.ESAlertQueryParams) ([]entities.ESAlertHistory, int64, error) {
	var alerts []entities.ESAlertHistory
	var total int64

	// 構建基礎查詢
	query := `
		SELECT
			time, monitor_id, alert_type, severity, status, message,
			cluster_name, threshold_value, actual_value, resolved_at,
			resolved_by, resolution_note, acknowledged_at, acknowledged_by, metadata
		FROM es_alert_history
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// 時間範圍過濾
	if !params.StartTime.IsZero() {
		query += fmt.Sprintf(" AND time >= $%d", argIndex)
		args = append(args, params.StartTime)
		argIndex++
	}
	if !params.EndTime.IsZero() {
		query += fmt.Sprintf(" AND time <= $%d", argIndex)
		args = append(args, params.EndTime)
		argIndex++
	}

	// 狀態過濾
	if len(params.Status) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, pq.Array(params.Status))
		argIndex++
	}

	// 嚴重性過濾
	if len(params.Severity) > 0 {
		query += fmt.Sprintf(" AND severity = ANY($%d)", argIndex)
		args = append(args, pq.Array(params.Severity))
		argIndex++
	}

	// 監控器 ID 過濾
	if params.MonitorID > 0 {
		query += fmt.Sprintf(" AND monitor_id = $%d", argIndex)
		args = append(args, params.MonitorID)
		argIndex++
	}

	// 告警類型過濾
	if len(params.AlertType) > 0 {
		query += fmt.Sprintf(" AND alert_type = ANY($%d)", argIndex)
		args = append(args, pq.Array(params.AlertType))
		argIndex++
	}

	// 計算總數
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_query", query)
	err := global.TimescaleDB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// 排序
	query += " ORDER BY time DESC"

	// 分頁
	if params.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, params.PageSize, (params.Page-1)*params.PageSize)
	}

	// 執行查詢
	rows, err := global.TimescaleDB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	// 解析結果
	for rows.Next() {
		var alert entities.ESAlertHistory
		var clusterName, resolvedBy, resolutionNote, acknowledgedBy, metadata sql.NullString
		err := rows.Scan(
			&alert.Time,
			&alert.MonitorID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Status,
			&alert.Message,
			&clusterName,
			&alert.ThresholdValue,
			&alert.ActualValue,
			&alert.ResolvedAt,
			&resolvedBy,
			&resolutionNote,
			&alert.AcknowledgedAt,
			&acknowledgedBy,
			&metadata,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alert: %w", err)
		}

		// 處理 NULL 值的字串欄位
		if clusterName.Valid {
			alert.ClusterName = clusterName.String
		}
		if resolvedBy.Valid {
			alert.ResolvedBy = resolvedBy.String
		}
		if resolutionNote.Valid {
			alert.ResolutionNote = resolutionNote.String
		}
		if acknowledgedBy.Valid {
			alert.AcknowledgedBy = acknowledgedBy.String
		}
		if metadata.Valid {
			alert.Metadata = metadata.String
		}

		alerts = append(alerts, alert)
	}

	return alerts, total, nil
}

// GetAlertByID 根據 ID 獲取單一告警
func (s *ESAlertService) GetAlertByID(monitorID int, alertTime time.Time) (*entities.ESAlertHistory, error) {
	var alert entities.ESAlertHistory
	var clusterName, resolvedBy, resolutionNote, acknowledgedBy, metadata sql.NullString

	query := `
		SELECT
			time, monitor_id, alert_type, severity, status, message,
			cluster_name, threshold_value, actual_value, resolved_at,
			resolved_by, resolution_note, acknowledged_at, acknowledged_by, metadata
		FROM es_alert_history
		WHERE monitor_id = $1 AND time = $2
	`

	err := global.TimescaleDB.QueryRow(query, monitorID, alertTime).Scan(
		&alert.Time,
		&alert.MonitorID,
		&alert.AlertType,
		&alert.Severity,
		&alert.Status,
		&alert.Message,
		&clusterName,
		&alert.ThresholdValue,
		&alert.ActualValue,
		&alert.ResolvedAt,
		&resolvedBy,
		&resolutionNote,
		&alert.AcknowledgedAt,
		&acknowledgedBy,
		&metadata,
	)

	if err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	// 處理 NULL 值的字串欄位
	if clusterName.Valid {
		alert.ClusterName = clusterName.String
	}
	if resolvedBy.Valid {
		alert.ResolvedBy = resolvedBy.String
	}
	if resolutionNote.Valid {
		alert.ResolutionNote = resolutionNote.String
	}
	if acknowledgedBy.Valid {
		alert.AcknowledgedBy = acknowledgedBy.String
	}
	if metadata.Valid {
		alert.Metadata = metadata.String
	}

	return &alert, nil
}

// ResolveAlert 標記告警為已解決
func (s *ESAlertService) ResolveAlert(monitorID int, alertTime time.Time, resolvedBy, resolutionNote string) error {
	query := `
		UPDATE es_alert_history
		SET
			status = 'resolved',
			resolved_at = $1,
			resolved_by = $2,
			resolution_note = $3
		WHERE monitor_id = $4 AND time = $5 AND status != 'resolved'
	`

	result, err := global.TimescaleDB.Exec(
		query,
		time.Now(),
		resolvedBy,
		resolutionNote,
		monitorID,
		alertTime,
	)

	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}

// AcknowledgeAlert 確認告警
func (s *ESAlertService) AcknowledgeAlert(monitorID int, alertTime time.Time, acknowledgedBy string) error {
	query := `
		UPDATE es_alert_history
		SET
			acknowledged_at = $1,
			acknowledged_by = $2
		WHERE monitor_id = $3 AND time = $4
	`

	result, err := global.TimescaleDB.Exec(
		query,
		time.Now(),
		acknowledgedBy,
		monitorID,
		alertTime,
	)

	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// GetAlertStatistics 獲取告警統計資料
func (s *ESAlertService) GetAlertStatistics(startTime, endTime time.Time) (*models.ESAlertStatistics, error) {
	stats := &models.ESAlertStatistics{}

	// 總告警數
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'high' THEN 1 END) as high,
			COUNT(CASE WHEN severity = 'medium' THEN 1 END) as medium,
			COUNT(CASE WHEN severity = 'low' THEN 1 END) as low
		FROM es_alert_history
		WHERE time BETWEEN $1 AND $2
	`

	err := global.TimescaleDB.QueryRow(query, startTime, endTime).Scan(
		&stats.Total,
		&stats.Active,
		&stats.Resolved,
		&stats.Critical,
		&stats.High,
		&stats.Medium,
		&stats.Low,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get alert statistics: %w", err)
	}

	return stats, nil
}

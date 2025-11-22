package services

import (
	"crypto/tls"
	"fmt"
	"log-detect/entities"
	"log-detect/global"
	"log-detect/log"
	"net/http"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
)

// ESConnectionManager ES 連線管理器（單例模式）
type ESConnectionManager struct {
	mu            sync.RWMutex
	clients       map[int]*elasticsearch.Client // connectionID -> client
	defaultClient *elasticsearch.Client
	defaultConnID int
	initialized   bool
}

var (
	connectionManager     *ESConnectionManager
	connectionManagerOnce sync.Once
)

// GetESConnectionManager 取得連線管理器單例
func GetESConnectionManager() *ESConnectionManager {
	connectionManagerOnce.Do(func() {
		connectionManager = &ESConnectionManager{
			clients:     make(map[int]*elasticsearch.Client),
			initialized: false,
		}
	})
	return connectionManager
}

// Initialize 初始化連線管理器（從資料庫載入所有連線）
func (m *ESConnectionManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		log.Logrecord_no_rotate("INFO", "ES Connection Manager already initialized, skipping...")
		return nil
	}

	log.Logrecord_no_rotate("INFO", "Initializing ES Connection Manager...")

	var connections []entities.ESConnection
	if err := global.Mysql.Where("deleted_at IS NULL").Find(&connections).Error; err != nil {
		return fmt.Errorf("failed to load ES connections from database: %w", err)
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Found %d ES connection(s) in database", len(connections)))

	successCount := 0
	for _, conn := range connections {
		client, err := m.createClient(&conn)
		if err != nil {
			log.Logrecord_no_rotate("WARNING", fmt.Sprintf("Failed to create client for connection '%s' (ID: %d): %v", conn.Name, conn.ID, err))
			continue
		}

		m.clients[conn.ID] = client
		successCount++

		if conn.IsDefault {
			m.defaultClient = client
			m.defaultConnID = conn.ID
			log.Logrecord_no_rotate("INFO", fmt.Sprintf("Set default ES connection: '%s' (ID: %d)", conn.Name, conn.ID))
		}
	}

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Successfully initialized %d/%d ES connection(s)", successCount, len(connections)))

	// 如果沒有預設連線，嘗試從 setting.yml 載入（向後兼容）
	if m.defaultClient == nil {
		log.Logrecord_no_rotate("WARNING", "No default ES connection found in database, trying to load from setting.yml...")
		if err := m.loadFromConfig(); err != nil {
			return fmt.Errorf("no default connection found and failed to load from config: %w", err)
		}
		log.Logrecord_no_rotate("INFO", "Successfully loaded default ES connection from setting.yml")
	}

	m.initialized = true
	log.Logrecord_no_rotate("INFO", "ES Connection Manager initialized successfully")
	return nil
}

// GetClient 根據連線 ID 取得客戶端
func (m *ESConnectionManager) GetClient(connectionID int) (*elasticsearch.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil, fmt.Errorf("ES Connection Manager not initialized")
	}

	client, exists := m.clients[connectionID]
	if !exists {
		return nil, fmt.Errorf("ES connection %d not found or failed to initialize", connectionID)
	}

	return client, nil
}

// GetClientForIndex 為指定的 Index 取得對應的 ES 客戶端
func (m *ESConnectionManager) GetClientForIndex(indexID int) (*elasticsearch.Client, error) {
	var index entities.Index
	if err := global.Mysql.Preload("ESConnection").First(&index, indexID).Error; err != nil {
		return nil, fmt.Errorf("failed to load index %d: %w", indexID, err)
	}

	// 如果 Index 指定了 ES 連線，使用該連線
	if index.ESConnectionID != nil {
		log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Index %d using ES connection %d", indexID, *index.ESConnectionID))
		return m.GetClient(*index.ESConnectionID)
	}

	// 否則使用預設連線（向後兼容）
	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Index %d using default ES connection", indexID))
	return m.GetDefaultClient(), nil
}

// GetClientForMonitor 為指定的 ElasticsearchMonitor 取得對應的 ES 客戶端（Phase 4 使用）
func (m *ESConnectionManager) GetClientForMonitor(monitorID int) (*elasticsearch.Client, error) {
	var monitor entities.ElasticsearchMonitor
	if err := global.Mysql.Preload("ESConnection").First(&monitor, monitorID).Error; err != nil {
		return nil, fmt.Errorf("failed to load monitor %d: %w", monitorID, err)
	}

	// 如果 Monitor 指定了 ES 連線，使用該連線（複用 indices 的連線）
	if monitor.ESConnectionID != nil {
		log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d reusing ES connection %d", monitorID, *monitor.ESConnectionID))
		return m.GetClient(*monitor.ESConnectionID)
	}

	// 否則返回 nil，表示應該使用 Monitor 自己的 host/port 配置
	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Monitor %d using its own host/port configuration", monitorID))
	return nil, nil
}

// GetDefaultClient 取得預設客戶端
func (m *ESConnectionManager) GetDefaultClient() *elasticsearch.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultClient == nil {
		log.Logrecord_no_rotate("WARNING", "Default ES client is nil, this should not happen after initialization")
	}

	return m.defaultClient
}

// ReloadConnection 重新載入指定連線（配置變更時調用）
func (m *ESConnectionManager) ReloadConnection(connectionID int) error {
	var conn entities.ESConnection
	if err := global.Mysql.Where("id = ? AND deleted_at IS NULL", connectionID).First(&conn).Error; err != nil {
		return fmt.Errorf("failed to load connection %d: %w", connectionID, err)
	}

	client, err := m.createClient(&conn)
	if err != nil {
		return fmt.Errorf("failed to create client for connection %d: %w", connectionID, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新客戶端
	m.clients[connectionID] = client

	// 如果是預設連線，更新預設客戶端
	if conn.IsDefault {
		m.defaultClient = client
		m.defaultConnID = connectionID
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Reloaded default ES connection: '%s' (ID: %d)", conn.Name, conn.ID))
	} else {
		log.Logrecord_no_rotate("INFO", fmt.Sprintf("Reloaded ES connection: '%s' (ID: %d)", conn.Name, conn.ID))
	}

	return nil
}

// ReloadAllConnections 重新載入所有連線（批量更新時調用）
func (m *ESConnectionManager) ReloadAllConnections() error {
	log.Logrecord_no_rotate("INFO", "Reloading all ES connections...")

	m.mu.Lock()
	m.initialized = false
	m.clients = make(map[int]*elasticsearch.Client)
	m.defaultClient = nil
	m.defaultConnID = 0
	m.mu.Unlock()

	return m.Initialize()
}

// RemoveConnection 移除指定連線（刪除時調用）
func (m *ESConnectionManager) RemoveConnection(connectionID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, connectionID)
	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Removed ES connection %d from manager", connectionID))

	// 如果刪除的是預設連線，清空預設連線
	if m.defaultConnID == connectionID {
		m.defaultClient = nil
		m.defaultConnID = 0
		log.Logrecord_no_rotate("WARNING", "Default ES connection was removed, please set a new default connection")
	}
}

// createClient 建立 ES 客戶端（私有方法）
func (m *ESConnectionManager) createClient(conn *entities.ESConnection) (*elasticsearch.Client, error) {
	// 驗證連線配置
	if err := conn.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	cfg := elasticsearch.Config{
		Addresses: conn.GetURLs(),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳過證書驗證（生產環境建議設為 false 並配置正確的證書）
			},
		},
	}

	// 如果啟用認證，設定用戶名和密碼
	if conn.EnableAuth {
		cfg.Username = conn.Username
		cfg.Password = conn.Password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 測試連線（optional，可註解掉以加速啟動）
	// res, err := client.Info()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to connect to ES: %w", err)
	// }
	// defer res.Body.Close()

	log.Logrecord_no_rotate("DEBUG", fmt.Sprintf("Created ES client for connection '%s': %s", conn.Name, conn.GetURL()))

	return client, nil
}

// loadFromConfig 從 setting.yml 載入預設連線（向後兼容）
func (m *ESConnectionManager) loadFromConfig() error {
	if global.EnvConfig == nil || len(global.EnvConfig.ES.URL) == 0 {
		return fmt.Errorf("ES configuration not found in setting.yml")
	}

	cfg := elasticsearch.Config{
		Addresses: global.EnvConfig.ES.URL,
		Username:  global.EnvConfig.ES.SourceAccount,
		Password:  global.EnvConfig.ES.SourcePassword,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create ES client from config: %w", err)
	}

	m.defaultClient = client
	m.defaultConnID = 0 // 0 表示來自 config，非資料庫

	log.Logrecord_no_rotate("INFO", fmt.Sprintf("Loaded default ES connection from config: %v", global.EnvConfig.ES.URL))

	return nil
}

// GetConnectionStats 取得連線統計資訊（用於監控和診斷）
func (m *ESConnectionManager) GetConnectionStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"initialized":       m.initialized,
		"total_connections": len(m.clients),
		"default_conn_id":   m.defaultConnID,
		"has_default":       m.defaultClient != nil,
	}
}

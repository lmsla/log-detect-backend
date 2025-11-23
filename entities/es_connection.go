package entities

import (
	"fmt"
	"log-detect/models"
)

// ESConnection Elasticsearch 連線配置（基礎實體）
type ESConnection struct {
	models.Common
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name" form:"name"`
	Host        string `gorm:"type:varchar(255);not null" json:"host" form:"host"`
	Port        int    `gorm:"type:int;not null;default:9200" json:"port" form:"port"`
	Username    string `gorm:"type:varchar(100)" json:"username" form:"username"`
	Password    string `gorm:"type:varchar(255)" json:"password,omitempty" form:"password"` // omitempty 避免在 JSON 中返回密碼
	EnableAuth  bool   `gorm:"type:tinyint(1);default:0" json:"enable_auth" form:"enable_auth"`
	UseTLS      bool   `gorm:"type:tinyint(1);default:1" json:"use_tls" form:"use_tls"`
	IsDefault   bool   `gorm:"type:tinyint(1);default:0" json:"is_default" form:"is_default"`
	Description string `gorm:"type:text" json:"description" form:"description"`
}

// TableName 指定表名
func (ESConnection) TableName() string {
	return "es_connections"
}

// GetURL 取得完整的 ES URL
func (c *ESConnection) GetURL() string {
	protocol := "http"
	if c.UseTLS {
		protocol = "https"
	}
	// 清理 Host 中可能存在的協議前綴
	host := c.CleanHost()
	return fmt.Sprintf("%s://%s:%d", protocol, host, c.Port)
}

// CleanHost 清理 Host 欄位中的協議前綴
func (c *ESConnection) CleanHost() string {
	host := c.Host
	// 移除各種可能的協議前綴
	prefixes := []string{"https://", "http://", "https//", "http//", "https:", "http:"}
	for _, prefix := range prefixes {
		if len(host) > len(prefix) && host[:len(prefix)] == prefix {
			host = host[len(prefix):]
			break
		}
	}
	return host
}

// GetURLs 取得 URL 陣列（用於 elasticsearch.Config.Addresses）
func (c *ESConnection) GetURLs() []string {
	return []string{c.GetURL()}
}

// Validate 驗證連線配置的有效性
func (c *ESConnection) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("連線名稱不能為空")
	}
	if c.Host == "" {
		return fmt.Errorf("主機地址不能為空")
	}
	// 自動清理 Host 中的協議前綴
	c.Host = c.CleanHost()
	if c.Host == "" {
		return fmt.Errorf("主機地址不能為空")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("端口必須在 1-65535 之間")
	}
	if c.EnableAuth && (c.Username == "" || c.Password == "") {
		return fmt.Errorf("啟用認證時，用戶名和密碼不能為空")
	}
	return nil
}

// MaskPassword 隱藏密碼（用於日誌記錄）
func (c *ESConnection) MaskPassword() *ESConnection {
	masked := *c
	if masked.Password != "" {
		masked.Password = "********"
	}
	return &masked
}

// ESConnectionSummary ES 連線摘要（用於列表顯示，不包含敏感資訊）
type ESConnectionSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	EnableAuth  bool   `json:"enable_auth"`
	UseTLS      bool   `json:"use_tls"`
	IsDefault   bool   `json:"is_default"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"` // online, offline, unknown（連線測試結果）
}

// ToSummary 轉換為摘要格式（移除敏感資訊）
func (c *ESConnection) ToSummary() ESConnectionSummary {
	return ESConnectionSummary{
		ID:          c.ID,
		Name:        c.Name,
		Host:        c.Host,
		Port:        c.Port,
		EnableAuth:  c.EnableAuth,
		UseTLS:      c.UseTLS,
		IsDefault:   c.IsDefault,
		Description: c.Description,
	}
}

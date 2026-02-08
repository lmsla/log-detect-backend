package structs

// FeaturesConfig 功能模組開關
type FeaturesConfig struct {
	TimescaleDB  bool `mapstructure:"timescaledb"`
	ESMonitoring bool `mapstructure:"es_monitoring"`
	Dashboard    bool `mapstructure:"dashboard"`
	Auth         bool `mapstructure:"auth"`
	History      bool `mapstructure:"history"`
}

// YMLConfig config.yml 擴充格式的根結構
type YMLConfig struct {
	ESConnections []YMLESConnection `yaml:"es_connections" mapstructure:"es_connections"`
	Targets       []YMLTarget       `yaml:"targets" mapstructure:"targets"`
	Devices       []YMLDeviceGroup  `yaml:"devices" mapstructure:"devices"`
}

// YMLESConnection ES 連線配置（YML 格式）
type YMLESConnection struct {
	Name        string `yaml:"name" mapstructure:"name"`
	Host        string `yaml:"host" mapstructure:"host"`
	Port        int    `yaml:"port" mapstructure:"port"`
	Username    string `yaml:"username" mapstructure:"username"`
	Password    string `yaml:"password" mapstructure:"password"`
	EnableAuth  bool   `yaml:"enable_auth" mapstructure:"enable_auth"`
	UseTLS      bool   `yaml:"use_tls" mapstructure:"use_tls"`
	IsDefault   bool   `yaml:"is_default" mapstructure:"is_default"`
	Description string `yaml:"description" mapstructure:"description"`
}

// YMLTarget 監控目標配置（YML 格式）
type YMLTarget struct {
	Subject  string     `yaml:"subject" mapstructure:"subject"`
	Receiver []string   `yaml:"receiver" mapstructure:"receiver"`
	Enable   bool       `yaml:"enable" mapstructure:"enable"`
	Indices  []YMLIndex `yaml:"indices" mapstructure:"indices"`
}

// YMLIndex 索引配置（YML 格式）
type YMLIndex struct {
	Index        string `yaml:"index" mapstructure:"index"`
	Logname      string `yaml:"logname" mapstructure:"logname"`
	DeviceGroup  string `yaml:"device_group" mapstructure:"device_group"`
	Period       string `yaml:"period" mapstructure:"period"`
	Unit         int    `yaml:"unit" mapstructure:"unit"`
	Field        string `yaml:"field" mapstructure:"field"`
	ESConnection string `yaml:"es_connection" mapstructure:"es_connection"`
}

// YMLDeviceGroup 裝置群組配置（YML 格式）
type YMLDeviceGroup struct {
	DeviceGroup string          `yaml:"device_group" mapstructure:"device_group"`
	Names       []YMLDeviceItem `yaml:"names" mapstructure:"names"`
}

// YMLDeviceItem 單一裝置配置（支援 HA 群組）
type YMLDeviceItem struct {
	Name    string `yaml:"name" mapstructure:"name"`
	HAGroup string `yaml:"ha_group" mapstructure:"ha_group"`
}

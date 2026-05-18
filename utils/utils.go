package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log-detect/global"
	"log-detect/structs"
	"os"
	"strings"
)

func LoadEnvironment() {
	loadSettingFile()
	viperSettingToModel()
	loadConfigFile()
	loadDevicesFile()
	viperconfigToModel()
}

func loadConfigFile() {
	configViper := viper.New()
	configViper.SetConfigName("config")
	configViper.SetConfigType("yml")
	configViper.AddConfigPath(".")

	if err := configViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("未發現 config.yml，跳過配置載入（適用於 API 模式）")
			global.TargetStruct = &structs.TargetStruct{}
			global.SetYMLConfig(&structs.YMLConfig{})
			return
		}
		panic(err)
	}

	// 保留原有的 TargetStruct 解析（向後相容）
	var c structs.TargetStruct
	if err := configViper.Unmarshal(&c); err != nil {
		panic(err)
	}
	global.TargetStruct = &c

	// 解析擴充格式到 YMLConfig（用於 YML-to-DB 同步）
	var ymlConfig structs.YMLConfig
	if err := configViper.Unmarshal(&ymlConfig); err != nil {
		fmt.Printf("Warning: could not unmarshal expanded config.yml: %v\n", err)
	}
	global.SetYMLConfig(&ymlConfig)
}

// loadDevicesFile 載入獨立的 devices.yml 裝置配置檔
// 如果檔案不存在則跳過（不影響啟動）
func loadDevicesFile() {
	devicesConfig, _, err := ReadDevicesFileConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("未發現 devices.yml，跳過裝置配置載入")
			return
		}
		fmt.Printf("Warning: 讀取 devices.yml 失敗: %v\n", err)
		return
	}

	baseConfig := cloneYMLConfig(global.GetYMLConfig())
	baseConfig.Devices = devicesConfig.Devices
	baseConfig.DisabledDevices = devicesConfig.DisabledDevices
	global.SetYMLConfig(baseConfig)

	if global.EnvConfig != nil && global.EnvConfig.ConfigSource == "api" {
		fmt.Printf("偵測到 devices.yml：已載入 %d 個裝置群組、%d 個停用群組至記憶體（API 模式，不會同步至 DB）\n",
			len(devicesConfig.Devices), len(devicesConfig.DisabledDevices))
		return
	}
	fmt.Printf("已從 devices.yml 載入 %d 個裝置群組、%d 個停用群組\n",
		len(devicesConfig.Devices), len(devicesConfig.DisabledDevices))
}

func ReadDevicesFileConfig() (*structs.YMLConfig, []byte, error) {
	data, err := os.ReadFile("devices.yml")
	if err != nil {
		return nil, nil, err
	}

	devicesConfig := &structs.YMLConfig{}
	if err := yaml.Unmarshal(data, devicesConfig); err != nil {
		fmt.Printf("Warning: 解析 devices.yml 失敗: %v\n", err)
		return nil, nil, err
	}

	return devicesConfig, data, nil
}

func cloneYMLConfig(cfg *structs.YMLConfig) *structs.YMLConfig {
	if cfg == nil {
		return &structs.YMLConfig{}
	}
	cloned := *cfg
	cloned.ESConnections = append([]structs.YMLESConnection(nil), cfg.ESConnections...)
	cloned.Targets = append([]structs.YMLTarget(nil), cfg.Targets...)
	cloned.Devices = append([]structs.YMLDeviceGroup(nil), cfg.Devices...)
	cloned.DisabledDevices = append([]structs.YMLDisabledDeviceGroup(nil), cfg.DisabledDevices...)
	return &cloned
}

func viperconfigToModel() {
	// var c structs.ActionStruct
	// c.Actions = viper.GetStringSlice("actions")
}

func loadSettingFile() {
	viper.SetConfigName("setting")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("沒有發現 setting.yml，改抓取環境變數")
			viper.AutomaticEnv()
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		} else {
			// 有找到 config.yml 但是發生了其他未知的錯誤
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

func viperSettingToModel() {
	var config structs.EnviromentModel
	// var action structs.Action

	/* sso */
	config.SSO.URL = viper.GetString("sso.url")
	config.SSO.Realm = viper.GetString("sso.realm")
	config.SSO.Username = viper.GetString("sso.username")
	config.SSO.Password = viper.GetString("sso.password")
	config.SSO.ClientID = viper.GetString("sso.client_id")
	config.SSO.AdminRole = viper.GetString("sso.admin_role")
	config.SSO.UserRole = viper.GetString("sso.user_role")
	config.SSO.DefaultSetting.Username = viper.GetString("sso.default_setting.username")
	config.SSO.DefaultSetting.Password = viper.GetString("sso.default_setting.password")
	config.SSO.DefaultSetting.Role = viper.GetString("sso.default_setting.role")
	config.SSO.DefaultSetting.Group = viper.GetString("sso.default_setting.group")

	// SQL Database
	config.Database.Client = viper.GetString("database.client")
	config.Database.Host = viper.GetString("database.host")
	config.Database.User = viper.GetString("database.user")
	config.Database.Password = viper.GetString("database.password")
	config.Database.Db = viper.GetString("database.name")
	config.Database.MaxIdle = uint(viper.GetInt("database.max_idle"))
	config.Database.MaxOpenConn = uint(viper.GetInt("database.max_open_conn"))
	config.Database.MaxLifeTime = viper.GetString("database.max_life_time")
	config.Database.Params = viper.GetString("database.params")
	config.Database.Port = viper.GetString("database.port")
	config.Database.LogEnable = viper.GetInt("database.log_enable")
	config.Database.Migration = viper.GetBool("database.migration")

	// TimescaleDB
	config.Timescale.Host = viper.GetString("timescale.host")
	config.Timescale.Port = viper.GetString("timescale.port")
	config.Timescale.User = viper.GetString("timescale.user")
	config.Timescale.Password = viper.GetString("timescale.password")
	config.Timescale.Db = viper.GetString("timescale.name")
	config.Timescale.MaxIdle = uint(viper.GetInt("timescale.max_idle"))

	config.Timescale.MaxOpenConn = uint(viper.GetInt("timescale.max_open_conn"))
	config.Timescale.MaxLifeTime = viper.GetString("timescale.max_life_time")
	config.Timescale.SSLMode = viper.GetString("timescale.sslmode")

	// BatchWriter
	config.BatchWriter.Enabled = viper.GetBool("batch_writer.enabled")

	config.BatchWriter.BatchSize = viper.GetInt("batch_writer.batch_size")
	config.BatchWriter.FlushInterval = viper.GetString("batch_writer.flush_interval")

	// YML Reload
	config.YMLReload.DevicesEnabled = viper.GetBool("yml_reload.devices_enabled")
	config.YMLReload.DevicesInterval = viper.GetString("yml_reload.devices_interval")

	config.Cors.Allow.Headers = viper.GetStringSlice("cors.allow.headers")

	config.Server.Mode = viper.GetString("server.mode")
	config.Server.Port = viper.GetString("server.port")

	//// ES
	config.ES.URL = viper.GetStringSlice("es.url")
	config.ES.SourceAccount = viper.GetString("es.sourceAccount")
	config.ES.SourcePassword = viper.GetString("es.sourcePassword")

	//// Path
	config.Path.Log_record = viper.GetString("path.log_record")

	/// Email
	config.Email.User = viper.GetString("email.user")
	config.Email.Password = viper.GetString("email.password")
	config.Email.Sender = viper.GetString("email.sender")
	config.Email.Host = viper.GetString("email.host")
	config.Email.Port = viper.GetString("email.port")
	config.Email.SMTP = viper.GetStringSlice("email.smtp")
	config.Email.AuthType = viper.GetString("email.auth_type")
	config.Email.DisableTLS = viper.GetBool("email.disable_tls")
	config.Email.Auth = viper.GetBool("email.auth")

	// Features（功能模組開關）
	config.Features.TimescaleDB = viper.GetBool("features.timescaledb")
	config.Features.ESMonitoring = viper.GetBool("features.es_monitoring")
	config.Features.Dashboard = viper.GetBool("features.dashboard")
	config.Features.Auth = viper.GetBool("features.auth")

	// ConfigSource（配置來源）
	config.ConfigSource = viper.GetString("config_source")

	global.EnvConfig = &config
}

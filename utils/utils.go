package utils

import (
	"log-detect/global"
	"log-detect/structs"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

func LoadEnvironment() {
	loadSettingFile()
	viperSettingToModel()
	loadConfigFile()
	viperconfigToModel()
}

func loadConfigFile() {
	// var configViperConfig = viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	//读取配置文件内容
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	var c structs.TargetStruct
	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}


	global.TargetStruct = &c
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
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
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

	global.EnvConfig = &config
	// global.Action = &action
}

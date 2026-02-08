package structs

type EnviromentModel struct {
	Database     database
	Timescale    timescale    // TimescaleDB 配置
	BatchWriter  batchWriter  // 批量寫入配置
	Server       server
	ES           es
	LIST         list
	Email        email
	INFORMATION  information
	Path         path
	Cors         corsModel
	SSO          sso
	Features     FeaturesConfig // 功能模組開關
	ConfigSource string         // 配置來源："yml" 或 "api"
}

type sso struct {
	URL            string `json:"url"`
	Realm          string `json:"realm"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	ClientID       string `json:"client_id"`
	AdminRole      string `json:"admin_role"`
	UserRole       string `json:"user_role"`
	DefaultSetting struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Group    string `json:"group"`
		Role     string `json:"role"`
	}
}

type email struct {
	User     string
	Password string
	Host     string
	Port     string
	Sender   string
	Auth     bool
	SMTP     []string
	AuthType string
	DisableTLS bool
}

type path struct {
	Log_record string
}

type database struct {
	Client      string
	MaxIdle     uint
	MaxLifeTime string
	MaxOpenConn uint
	User        string
	Password    string
	Host        string
	Db          string
	Params      string
	Port        string
	LogEnable   int
	Migration   bool
}

type corsModel struct {
	Allow corsAllowModel
}

type corsAllowModel struct {
	Headers []string
}

type server struct {
	Port string
	Mode string
}

// TimescaleDB 配置結構
type timescale struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Db          string `mapstructure:"name"`
	MaxIdle     uint   `mapstructure:"max_idle"`
	MaxLifeTime string `mapstructure:"max_life_time"`
	MaxOpenConn uint   `mapstructure:"max_open_conn"`
	SSLMode     string `mapstructure:"sslmode"`
}

// 批量寫入配置結構
type batchWriter struct {
	Enabled       bool   `mapstructure:"enabled"`
	BatchSize     int    `mapstructure:"batch_size"`
	FlushInterval string `mapstructure:"flush_interval"`
}

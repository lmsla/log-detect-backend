package structs

type EnviromentModel struct {
	Database    database
	Server      server
	ES          es
	LIST        list
	Email       email
	INFORMATION information
	Path        path
	Cors        corsModel
	SSO         sso
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

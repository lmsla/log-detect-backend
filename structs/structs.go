package structs

type TargetStruct struct {
	Targets []Receiverlist `yaml:"targets"`
	// Targets   []Index        `yaml:"targets"`
}

type Receiverlist struct {
	Receiver []string    `yaml:"receiver"`
	Subject  string      `yaml:"subject"`
	Indices  []Indexlist `yaml:"indices"`
}

type Indexlist struct {
	Index   string `yaml:"index"`
	Logname string `yaml:"logname"`
	Period  string `yaml:"period"`
	Unit    int    `yaml:"unit"`
	Field   string `yaml:"field"`
}

type information struct {
	CaPath       string
	Logdir       string
	Period       string
	Execute_cron bool
}

type es struct {
	URL            []string
	SourceAccount  string
	SourcePassword string
}

type list struct {
	Iplist    string
	Totallist string
}


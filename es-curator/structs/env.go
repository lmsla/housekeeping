package structs

type EnviromentModel struct {
	ES es
	// LIST            list
	// NotifyIndexName string
	INFORMATION information
	Log         log
}

type information struct {
	CaPath      string
	LogPath     string
	Period      string
	ExecuteCron bool
	TestMode    bool
}

type es struct {
	URL            []string
	SourceAccount  string
	SourcePassword string
}

type log struct {
	Path                string
	MaxSize             int
	MaxBackups          int
	MaxAge              int
	Debug               bool
	ToES                bool
	HealthCheckInterval int
}

type ActionStruct struct {
	Actions  []Actiond `yaml:"actions"`
	CronTime string    `yaml:"cron_time"`
}

type Actiond struct {
	Action        string   `yaml:"action"`
	Description   string   `yaml:"description"`
	Options       Option   `yaml:"options,omitempty"`
	Filters       []Filter `yaml:"filters,omitempty"`
	ExecutePeriod string   `yaml:"execute_period,omitempty"`
}

type Filter struct {
	Filtertype string   `yaml:"filtertype"`
	Kind       string   `yaml:"kind,omitempty"`
	Value      []string `yaml:"value,omitempty"`
	// Value      interface{} `yaml:"value,omitempty"`
	Source     string `yaml:"source,omitempty"`
	RangeFrom  int    `yaml:"range_from"`
	RangeTo    int    `yaml:"range_to"`
	Timestring string `yaml:"timestring,omitempty"`
	Unit       string `yaml:"unit,omitempty"`
	UnitCount  int    `yaml:"unit_count,omitempty"`
	Direction  string `yaml:"direction"`
	DiskSpace  int    `yaml:"disk_space"`
	UpperLimit int    `yaml:"upper_limit,omitempty"`
	LowerLimit int    `yaml:"lower_limit,omitempty"`
}

type Option struct {
	DisableAction     bool   `yaml:"disable_action"`
	WaitForCompletion bool   `yaml:"wait_for_completion"`
	Key               string `yaml:"key"`
	Value             string `yaml:"value"`
	AllocationType    string `yaml:"allocation_type"`
	MaxNumSegment     int    `yaml:"max_num_segment"`
	Delay             int    `yaml:"delay"`
	TimeoutOverride   int    `yaml:"TimeoutOverride"`
	// Rollover 相關配置
	RolloverAlias     string `yaml:"rollover_alias"`        // 要 rollover 的 alias
	MaxSize           string `yaml:"max_size,omitempty"`     // 觸發 rollover 的最大大小 (如 "50gb")
	MaxDocs           int64  `yaml:"max_docs,omitempty"`     // 觸發 rollover 的最大文檔數
	MaxAge            string `yaml:"max_age,omitempty"`      // 觸發 rollover 的最大時間 (如 "30d")
	NewIndexName      string `yaml:"new_index_name,omitempty"` // 可選：指定新索引名稱模式
}

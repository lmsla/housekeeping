package structs

type EnviromentModel struct {
	ES              es
	LIST            list
	NotifyIndexName string
	INFORMATION     information
}

type ActionStruct struct {
	Actions   []Actiond `yaml:"actions"`
	Cron_Time string    `yaml:"cron_time"`
}

type Actiond struct {
	Action         string   `yaml:"action"`
	Description    string   `yaml:"description"`
	Options        Option   `yaml:"options,omitempty"`
	Filters        []Filter `yaml:"filters,omitempty"`
	Execute_Period string   `yaml:"execute_period,omitempty"`
}

type Filter struct {
	Filtertype string `yaml:"filtertype"`
	Kind       string `yaml:"kind,omitempty"`
	Value      string `yaml:"value,omitempty"`
	Source     string `yaml:"source,omitempty"`
	Range_From int64  `yaml:"range_from"`
	Range_To   int    `yaml:"range_to"`
	Timestring string `yaml:"timestring,omitempty"`
	Unit       string `yaml:"unit,omitempty"`
	Unit_count int    `yaml:"unit_count,omitempty"`
	Direction  string `yaml:"direction,omitempty"`
	Disk_space int    `yaml:"disk_space,omitempty"`
}

type Option struct {
	DisableAction     bool   `mapstructure:"disable_action"`
	WaitForCompletion bool   `mapstructure:"wait_for_completion"`
	Key               string `yaml:"key"`
	Value             string `yaml:"value"`
	AllocationType    string `mapstructure:"allocation_type"`
	MaxNumSegment     int    `mapstructure:"max_num_segment"`
	Delay             int    `yaml:"delay"`
	TimeoutOverride   int    `yaml:"TimeoutOverride"`
}

type delete_indices struct {
	Action  string
	Filters filters
}

type filters struct {
	Type_age     type_age
	Type_pattern type_pattern
}

type type_age struct {
	Source     string
	Direction  string
	Timestring string
	Unit       string
	Unit_count int
}

type type_pattern struct {
	Kind    string
	Value   string
	Exclude bool
}

type information struct {
	CaPath       string
	LogPath       string
	Period       string
	Execute_cron bool
	Test_mode bool
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

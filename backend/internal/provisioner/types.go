package provisioner

type VendorConfig struct {
	ID                  string   `yaml:"id"`
	Name                string   `yaml:"name"`
	StaticDir           string   `yaml:"static_dir"`            // Директория со статикой (относительно vendor.yaml)
	PhoneConfigFile     string   `yaml:"phone_config_file"`     // Шаблон имени файла конфига телефона (например: "{{account.mac_address}}.cfg")
	PhoneConfigTemplate string   `yaml:"phone_config_template"` // Путь к шаблону конфига телефона (относительно vendor.yaml)
	FeaturesFile        string   `yaml:"features_file"`         // Путь к файлу с описанием функций (относительно vendor.yaml)
	AccountsFile        string   `yaml:"accounts_file"`         // Путь к файлу с описанием аккаунтов (относительно vendor.yaml)
	KeyTypes            []string `yaml:"key_types"`             // Список типов кнопок

	// Внутренние поля
	Dir      string    `yaml:"-"`
	Features []Feature `yaml:"-" json:"features"`
	Accounts []Feature `yaml:"-" json:"accounts"`
}

type Feature struct {
	ID     string         `yaml:"id" json:"id"`
	Name   string         `yaml:"name" json:"name"`
	Params []FeatureParam `yaml:"params" json:"params"`
}

type FeatureParam struct {
	ID             string                 `yaml:"id" json:"id"`
	Label          string                 `yaml:"label" json:"label"`
	Type           string                 `yaml:"type" json:"type"`                         // string, select, etc.
	Value          string                 `yaml:"value,omitempty" json:"value,omitempty"`   // Fixed value if any
	ConfigTemplate string                 `yaml:"config_template" json:"config_template"`   // Template line
	Source         string                 `yaml:"source,omitempty" json:"source,omitempty"` // e.g. "lines" for select source
	Extra          map[string]interface{} `yaml:",inline" json:"extra"`                     // Capture any other fields
}

type DeviceModel struct {
	ID     string `yaml:"id" json:"id"`
	Vendor string `yaml:"vendor" json:"vendor"`
	Name   string `yaml:"name" json:"name"`
	Type   string `yaml:"type" json:"type"` // "phone", "gateway", "expansion-module"
	Image  string `yaml:"image" json:"image"`
	//	OwnKeys                   int                 `yaml:"own_keys" json:"own_keys"`
	OwnSoftKeys               int                 `yaml:"own_soft_keys" json:"own_soft_keys"`
	OwnHardKeys               int                 `yaml:"own_hard_keys" json:"own_hard_keys"`
	SupportedExpansionModules []string            `yaml:"supported_expansion_modules" json:"supported_expansion_modules"`
	MaximumExpansionModules   int                 `yaml:"maximum_expansion_modules" json:"maximum_expansion_modules"`
	MaxAccountLines           int                 `yaml:"max_account_lines" json:"max_account_lines"`
	LineNameFormat            string              `yaml:"line_name_format" json:"line_name_format"` // Regex or format string
	Keys                      []ModelKey          `yaml:"keys" json:"keys"`
	KeyTypes                  []KeyType           `yaml:"key_types" json:"key_types"`
	Settings                  []ModelSettingGroup `yaml:"settings" json:"settings"`
}

type KeyType struct {
	ID      string `yaml:"id" json:"id"`
	Verbose string `yaml:"verbose" json:"verbose"`
	Polygon string `yaml:"polygon" json:"polygon"`
	Image   string `yaml:"image" json:"image"`
}

type ModelKey struct {
	Index    int               `yaml:"index" json:"index"`
	Type     string            `yaml:"type" json:"type"`
	Account  int               `yaml:"account" json:"account"` // Default account index
	Label    string            `yaml:"label" json:"label"`
	X        int               `yaml:"x" json:"x"`
	Y        int               `yaml:"y" json:"y"`
	MyImage  string            `yaml:"my_image" json:"my_image"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

type ModelSettingGroup struct {
	Group  string       `yaml:"group" json:"group"`
	Params []ModelParam `yaml:"params" json:"params"`
}

type ModelParam struct {
	Key     string      `yaml:"key" json:"key"`
	Label   string      `yaml:"label" json:"label"`
	Type    string      `yaml:"type" json:"type"` // string, boolean, number, select
	Default interface{} `yaml:"default" json:"default"`
	Options []string    `yaml:"options,omitempty" json:"options,omitempty"` // For select type
}

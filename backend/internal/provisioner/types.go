package provisioner

type VendorConfig struct {
	ID                  string `yaml:"id"`
	Name                string `yaml:"name"`
	StaticDir           string `yaml:"static_dir"`            // Директория со статикой (относительно vendor.yaml)
	PhoneConfigFile     string `yaml:"phone_config_file"`     // Шаблон имени файла конфига телефона (например: "{{account.mac_address}}.cfg")
	PhoneConfigTemplate string `yaml:"phone_config_template"` // Путь к шаблону конфига телефона (относительно vendor.yaml)

	// Внутреннее поле для хранения полного пути к директории вендора
	Dir string `yaml:"-"`
}

type DeviceModel struct {
	ID                        string              `yaml:"id" json:"id"`
	Vendor                    string              `yaml:"vendor" json:"vendor"`
	Name                      string              `yaml:"name" json:"name"`
	Type                      string              `yaml:"type" json:"type"` // "phone" (default) or "gateway"
	Image                     string              `yaml:"image" json:"image"`
	Lines                     int                 `yaml:"lines" json:"lines"`
	SupportedExpansionModules []string            `yaml:"supported_expansion_modules" json:"supported_expansion_modules"`
	MaximumExpansionModules   int                 `yaml:"maximum_expansion_modules" json:"maximum_expansion_modules"`
	MaxAdditionalLines        int                 `yaml:"max_additional_lines" json:"max_additional_lines"`
	LineNameFormat            string              `yaml:"line_name_format" json:"line_name_format"` // Regex or format string
	Keys                      []ModelKey          `yaml:"keys" json:"keys"`
	Settings                  []ModelSettingGroup `yaml:"settings" json:"settings"`
}

type ModelKey struct {
	Index    int               `yaml:"index" json:"index"`
	Type     string            `yaml:"type" json:"type"`
	Label    string            `yaml:"label" json:"label"`
	X        int               `yaml:"x" json:"x"`
	Y        int               `yaml:"y" json:"y"`
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

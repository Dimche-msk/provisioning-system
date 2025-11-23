package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DomainSettings struct {
	Name      string            `yaml:"name"`
	DeployCmd string            `yaml:"deploy_cmd"`
	DeleteCmd string            `yaml:"delete_cmd"`
	Variables map[string]string `yaml:"variables"`
}

type SystemConfig struct {
	Server struct {
		Port            string `yaml:"port"`
		ServeConfigs    bool   `yaml:"serve_configs"`
		LogDeviceAccess string `yaml:"log_device_access"` // none, access, error, full
		LogFilePath     string `yaml:"log_file_path"`
	} `yaml:"server"`
	Auth struct {
		AdminUser     string `yaml:"admin_user"`
		AdminPassword string `yaml:"admin_password"`
		SecretKey     string `yaml:"secret_key"`
	} `yaml:"auth"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Defaults DomainSettings   `yaml:"defaults"`
	Domains  []DomainSettings `yaml:"domains"`
}

func LoadConfig(configDir string) (*SystemConfig, error) {
	configPath := filepath.Join(configDir, "provisioning-system.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg SystemConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Дефолтные значения
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8090"
	}
	if cfg.Auth.SecretKey == "" {
		cfg.Auth.SecretKey = "change-me-in-production"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "provisioning.db"
	}

	return &cfg, nil
}

// GetEffectiveDomainConfig возвращает настройки для указанного домена,
// объединяя их с дефолтными настройками.
func (cfg *SystemConfig) GetEffectiveDomainConfig(domainName string) DomainSettings {
	// 1. Начинаем с копии дефолтных настроек
	effective := DomainSettings{
		Name:      cfg.Defaults.Name,
		DeployCmd: cfg.Defaults.DeployCmd,
		DeleteCmd: cfg.Defaults.DeleteCmd,
		Variables: make(map[string]string),
	}
	// Копируем переменные
	for k, v := range cfg.Defaults.Variables {
		effective.Variables[k] = v
	}

	// Если запрошен дефолтный домен или пустой - возвращаем сразу
	if domainName == "" || domainName == cfg.Defaults.Name {
		return effective
	}

	// 2. Ищем запрошенный домен
	var targetDomain *DomainSettings
	for _, d := range cfg.Domains {
		if d.Name == domainName {
			targetDomain = &d
			break
		}
	}

	// Если домен не найден, возвращаем дефолтный (или можно вернуть ошибку, но пока так безопаснее)
	if targetDomain == nil {
		return effective
	}

	// 3. Переопределяем настройки
	effective.Name = targetDomain.Name
	if targetDomain.DeployCmd != "" {
		effective.DeployCmd = targetDomain.DeployCmd
	}
	if targetDomain.DeleteCmd != "" {
		effective.DeleteCmd = targetDomain.DeleteCmd
	}

	// 4. Переопределяем/добавляем переменные
	for k, v := range targetDomain.Variables {
		effective.Variables[k] = v
	}

	return effective
}

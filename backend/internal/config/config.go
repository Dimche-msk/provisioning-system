package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DomainSettings struct {
	Name                   string            `yaml:"name"`
	DeployCmd              string            `yaml:"deploy_cmd"`               // Legacy: single command
	DeleteCmd              string            `yaml:"delete_cmd"`               // Legacy: single command
	DeployCommands         []string          `yaml:"deploy_commands"`          // New: list of commands
	DeleteCommands         []string          `yaml:"delete_commands"`          // New: list of commands
	GenerateRandomPassword bool              `yaml:"generate_random_password"` // If true, generate random password for new phones
	Variables              map[string]string `yaml:"variables"`
}

type SystemConfig struct {
	Server struct {
		ListenAddress   string `yaml:"listen_address"`
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
		Path      string `yaml:"path"`
		BackupDir string `yaml:"backup_dir"`
	} `yaml:"database"`
	Domains []DomainSettings `yaml:"domains"`
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
	if cfg.Server.ListenAddress == "" {
		cfg.Server.ListenAddress = "0.0.0.0"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8090"
	}
	if cfg.Auth.SecretKey == "" {
		cfg.Auth.SecretKey = "change-me-in-production"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "provisioning.db"
	}
	if cfg.Database.BackupDir == "" {
		cfg.Database.BackupDir = "backups"
	}

	return &cfg, nil
}

// GetEffectiveDomainConfig возвращает настройки для указанного домена.
// Если домен не найден, возвращает настройки первого домена (дефолтного).
func (cfg *SystemConfig) GetEffectiveDomainConfig(domainName string) DomainSettings {
	// 1. Находим запрошенный домен
	var targetDomain *DomainSettings
	for _, d := range cfg.Domains {
		if d.Name == domainName {
			targetDomain = &d
			break
		}
	}

	// 2. Если не найден, берем первый (дефолтный)
	if targetDomain == nil && len(cfg.Domains) > 0 {
		targetDomain = &cfg.Domains[0]
	}

	// Если вообще нет доменов (что странно), возвращаем пустой
	if targetDomain == nil {
		return DomainSettings{Variables: make(map[string]string)}
	}

	// 3. Копируем настройки (чтобы не менять оригинал, если будем модифицировать)
	effective := DomainSettings{
		Name:                   targetDomain.Name,
		DeployCmd:              targetDomain.DeployCmd,
		DeleteCmd:              targetDomain.DeleteCmd,
		DeployCommands:         make([]string, len(targetDomain.DeployCommands)),
		DeleteCommands:         make([]string, len(targetDomain.DeleteCommands)),
		GenerateRandomPassword: targetDomain.GenerateRandomPassword,
		Variables:              make(map[string]string),
	}
	copy(effective.DeployCommands, targetDomain.DeployCommands)
	copy(effective.DeleteCommands, targetDomain.DeleteCommands)

	// Backward compatibility: if new list is empty but old string is set, use it
	if len(effective.DeployCommands) == 0 && effective.DeployCmd != "" {
		effective.DeployCommands = []string{effective.DeployCmd}
	}
	if len(effective.DeleteCommands) == 0 && effective.DeleteCmd != "" {
		effective.DeleteCommands = []string{effective.DeleteCmd}
	}

	for k, v := range targetDomain.Variables {
		effective.Variables[k] = v
	}

	return effective
}

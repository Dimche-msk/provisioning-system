package provisioner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"provisioning-system/internal/config"
	"provisioning-system/internal/models"

	"github.com/flosch/pongo2/v6"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	Config  *config.SystemConfig
	Vendors []VendorConfig
	Models  []DeviceModel
}

func NewManager(cfg *config.SystemConfig) *Manager {
	return &Manager{
		Config:  cfg,
		Vendors: []VendorConfig{},
		Models:  []DeviceModel{},
	}
}

// getVendorByID ищет вендора по ID
func (m *Manager) getVendorByID(id string) (VendorConfig, bool) {
	for _, v := range m.Vendors {
		if v.ID == id {
			return v, true
		}
	}
	return VendorConfig{}, false
}

// LoadModels сканирует директории models внутри каждого вендора
func (m *Manager) LoadModels() error {
	m.Models = []DeviceModel{}

	for _, vendor := range m.Vendors {
		modelsDir := filepath.Join(vendor.Dir, "models")

		// Check if models directory exists
		if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(modelsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read model config %s: %w", path, err)
			}

			var model DeviceModel
			if err := yaml.Unmarshal(data, &model); err != nil {
				return fmt.Errorf("failed to parse model config %s: %w", path, err)
			}

			// Set vendor ID programmatically
			model.Vendor = vendor.ID
			m.Models = append(m.Models, model)
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk models directory for vendor %s: %w", vendor.Name, err)
		}
	}

	return nil
}

// LoadVendors сканирует директорию vendors и загружает конфигурации
func (m *Manager) LoadVendors(vendorsDir string) error {
	entries, err := os.ReadDir(vendorsDir)
	if err != nil {
		return fmt.Errorf("failed to read vendors directory: %w", err)
	}

	m.Vendors = []VendorConfig{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		vendorDir := filepath.Join(vendorsDir, entry.Name())
		configFile := filepath.Join(vendorDir, "vendor.yaml")

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue // Пропускаем папки без конфига
		}

		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read vendor config %s: %w", configFile, err)
		}

		var vc VendorConfig
		if err := yaml.Unmarshal(data, &vc); err != nil {
			return fmt.Errorf("failed to parse vendor config %s: %w", configFile, err)
		}

		vc.Dir = vendorDir
		m.Vendors = append(m.Vendors, vc)
	}

	return nil
}

// GenerateConfigs генерирует конфигурационные файлы для всех доменов и вендоров
func (m *Manager) GenerateConfigs(outputDir string) error {
	// 1. Проходим по всем доменам (включая дефолтный, если он явно задан в Domains,
	// но обычно мы хотим сгенерировать для Default + всех остальных)

	// Соберем список всех доменов для генерации
	domainsToProcess := []string{m.Config.Defaults.Name}
	for _, d := range m.Config.Domains {
		if d.Name != m.Config.Defaults.Name {
			domainsToProcess = append(domainsToProcess, d.Name)
		}
	}

	for _, domainName := range domainsToProcess {
		domainConfig := m.Config.GetEffectiveDomainConfig(domainName)

		for _, vendor := range m.Vendors {
			if err := m.generateForVendor(outputDir, domainConfig, vendor); err != nil {
				return fmt.Errorf("failed to generate config for domain %s, vendor %s: %w", domainName, vendor.Name, err)
			}
		}
	}

	return nil
}

func (m *Manager) generateForVendor(outputDir string, domain config.DomainSettings, vendor VendorConfig) error {
	// Создаем выходную директорию: outputDir/domain/
	// Файлы всех вендоров кладем в корень домена (как на TFTP сервере)
	targetDir := filepath.Join(outputDir, domain.Name)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", targetDir, err)
	}

	// Проходим по всем файлам в директории вендора
	err := filepath.Walk(vendor.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Обрабатываем только .tpl файлы
		if filepath.Ext(path) != ".tpl" {
			return nil
		}

		// Читаем шаблон
		tpl, err := pongo2.FromFile(path)
		if err != nil {
			return fmt.Errorf("failed to load template %s: %w", path, err)
		}

		// Подготавливаем контекст
		ctx := pongo2.Context{}
		for k, v := range domain.Variables {
			ctx[k] = v
		}
		ctx["domain_name"] = domain.Name
		ctx["vendor_name"] = vendor.Name

		// Рендерим
		out, err := tpl.Execute(ctx)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", path, err)
		}

		// Определяем относительный путь
		relPath, err := filepath.Rel(vendor.Dir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Если файл лежит в папке templates, убираем её из пути
		// Чтобы aastra.cfg лежал в корне, а не в templates/aastra.cfg
		if strings.HasPrefix(relPath, "templates"+string(os.PathSeparator)) {
			relPath = strings.TrimPrefix(relPath, "templates"+string(os.PathSeparator))
		}

		// Убираем расширение .tpl
		outputRelPath := strings.TrimSuffix(relPath, ".tpl")
		targetFile := filepath.Join(targetDir, outputRelPath)

		// Создаем поддиректории если нужно
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", targetFile, err)
		}

		// Записываем файл
		if err := os.WriteFile(targetFile, []byte(out), 0644); err != nil {
			return fmt.Errorf("failed to write config file %s: %w", targetFile, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk vendor directory %s: %w", vendor.Dir, err)
	}

	// Копируем статику, если она задана
	if vendor.StaticDir != "" {
		staticPath := filepath.Join(vendor.Dir, vendor.StaticDir)
		if err := m.copyStaticFiles(staticPath, targetDir); err != nil {
			return fmt.Errorf("failed to copy static files for vendor %s: %w", vendor.Name, err)
		}
	}

	return nil
}

func (m *Manager) copyStaticFiles(srcDir, dstDir string) error {
	// Проверяем, существует ли исходная директория
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // Если статики нет, это не ошибка (хотя если она указана в конфиге, может и стоит варнинг кинуть)
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		targetFile := filepath.Join(dstDir, relPath)

		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetFile, data, 0644)
	})
}

// GeneratePhoneConfigs генерирует конфигурации для списка телефонов
func (m *Manager) GeneratePhoneConfigs(outputDir string, phones []models.Phone) error {
	for _, phone := range phones {
		mac := ""
		if phone.MacAddress != nil {
			mac = *phone.MacAddress
		}
		number := ""
		if phone.PhoneNumber != nil {
			number = *phone.PhoneNumber
		}

		vendor, ok := m.getVendorByID(phone.Vendor)
		if !ok {
			log.Printf("Warning: Vendor %s not found for phone %s", phone.Vendor, mac)
			continue
		}

		if vendor.PhoneConfigFile == "" || vendor.PhoneConfigTemplate == "" {
			continue // Вендор не поддерживает генерацию конфигов телефонов
		}

		// Парсим настройки аккаунта
		var settings map[string]interface{}
		if phone.AccountSettings != "" {
			if err := json.Unmarshal([]byte(phone.AccountSettings), &settings); err != nil {
				log.Printf("Warning: Failed to parse account settings for phone %s: %v", mac, err)
				settings = make(map[string]interface{})
			}
		} else {
			settings = make(map[string]interface{})
		}

		// Подготавливаем контекст домена (флатеним переменные)
		domainConfig := m.Config.GetEffectiveDomainConfig(phone.Domain)
		domainCtx := make(map[string]interface{})
		for k, v := range domainConfig.Variables {
			domainCtx[k] = v
		}
		domainCtx["name"] = domainConfig.Name

		// Подготавливаем контекст
		ctx := pongo2.Context{
			"account": map[string]interface{}{
				"id":           phone.ID,
				"domain":       phone.Domain,
				"vendor":       phone.Vendor,
				"mac_address":  mac,
				"phone_number": number,
				"ip_address":   phone.IPAddress,
				"caller_id":    phone.CallerID,
				"settings":     settings,
			},
			"domain": domainCtx,
		}

		// 1. Генерируем имя файла
		filenameTpl, err := pongo2.FromString(vendor.PhoneConfigFile)
		if err != nil {
			log.Printf("Warning: Failed to parse filename template for vendor %s: %v", vendor.Name, err)
			continue
		}
		filename, err := filenameTpl.Execute(ctx)
		if err != nil {
			log.Printf("Warning: Failed to execute filename template for phone %s: %v", mac, err)
			continue
		}

		// 2. Генерируем содержимое
		templatePath := filepath.Join(vendor.Dir, vendor.PhoneConfigTemplate)
		tpl, err := pongo2.FromFile(templatePath)
		if err != nil {
			log.Printf("Warning: Failed to load template %s: %v", templatePath, err)
			continue
		}
		content, err := tpl.Execute(ctx)
		if err != nil {
			log.Printf("Warning: Failed to execute template for phone %s: %v", mac, err)
			continue
		}

		// 3. Записываем файл
		targetDir := filepath.Join(outputDir, phone.Domain)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", targetDir, err)
		}
		targetFile := filepath.Join(targetDir, filename)
		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			log.Printf("Warning: Failed to write config file %s: %v", targetFile, err)
			continue
		}
	}
	return nil
}

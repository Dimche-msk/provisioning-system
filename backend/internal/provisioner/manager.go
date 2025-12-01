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
	// Disable Pongo2 caching to allow hot-reloading of templates
	pongo2.DefaultSet.Debug = true

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

// getModelByID ищет модель по ID
func (m *Manager) getModelByID(id string) (DeviceModel, bool) {
	for _, model := range m.Models {
		if model.ID == id {
			return model, true
		}
	}
	return DeviceModel{}, false
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

		// Load Features if defined
		if vc.FeaturesFile != "" {
			featuresPath := filepath.Join(vendorDir, vc.FeaturesFile)
			if _, err := os.Stat(featuresPath); err == nil {
				featuresData, err := os.ReadFile(featuresPath)
				if err != nil {
					log.Printf("Warning: Failed to read features file %s: %v", featuresPath, err)
				} else {
					var features []Feature
					if err := yaml.Unmarshal(featuresData, &features); err != nil {
						log.Printf("Warning: Failed to parse features file %s: %v", featuresPath, err)
					} else {
						vc.Features = features
					}
				}
			}
		}

		m.Vendors = append(m.Vendors, vc)
	}

	return nil
}

// GenerateConfigs генерирует конфигурационные файлы для всех доменов и вендоров
func (m *Manager) GenerateConfigs(outputDir string, phones []models.Phone) error {
	// 1. Prepare data for all domains
	var allDomains []map[string]interface{}

	// Group phones by domain
	phonesByDomain := make(map[string][]models.Phone)
	for _, p := range phones {
		phonesByDomain[p.Domain] = append(phonesByDomain[p.Domain], p)
	}

	for _, d := range m.Config.Domains {
		domainData := map[string]interface{}{
			"name":      d.Name,
			"variables": d.Variables,
			"phones":    phonesByDomain[d.Name],
		}
		allDomains = append(allDomains, domainData)
	}

	// 2. Generate configs for each domain
	for _, d := range m.Config.Domains {
		domainName := d.Name
		domainConfig := m.Config.GetEffectiveDomainConfig(domainName)

		// Filter phones for this domain (still needed for context["phones"] backward compatibility/convenience)
		domainPhones := phonesByDomain[domainName]

		for _, vendor := range m.Vendors {
			if err := m.generateForVendor(outputDir, domainConfig, vendor, domainPhones, allDomains); err != nil {
				return fmt.Errorf("failed to generate config for domain %s, vendor %s: %w", domainName, vendor.Name, err)
			}
		}
	}

	return nil
}

func (m *Manager) generateForVendor(outputDir string, domain config.DomainSettings, vendor VendorConfig, phones []models.Phone, allDomains []map[string]interface{}) error {
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
		ctx["phones"] = phones
		ctx["all_domains"] = allDomains

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
		// Аналогично для directory
		if strings.HasPrefix(relPath, "directory"+string(os.PathSeparator)) {
			relPath = strings.TrimPrefix(relPath, "directory"+string(os.PathSeparator))
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

// GenerateDirectories генерирует только файлы из папки directory для всех вендоров
func (m *Manager) GenerateDirectories(outputDir string, phones []models.Phone) error {
	// 1. Prepare data for all domains
	var allDomains []map[string]interface{}

	// Group phones by domain
	phonesByDomain := make(map[string][]models.Phone)
	for _, p := range phones {
		phonesByDomain[p.Domain] = append(phonesByDomain[p.Domain], p)
	}

	for _, d := range m.Config.Domains {
		domainData := map[string]interface{}{
			"name":      d.Name,
			"variables": d.Variables,
			"phones":    phonesByDomain[d.Name],
		}
		allDomains = append(allDomains, domainData)
	}

	// 2. Generate directories for each domain
	for _, d := range m.Config.Domains {
		domainName := d.Name
		domainConfig := m.Config.GetEffectiveDomainConfig(domainName)
		domainPhones := phonesByDomain[domainName]

		for _, vendor := range m.Vendors {
			// Check if directory folder exists
			dirPath := filepath.Join(vendor.Dir, "directory")
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				continue
			}

			// We use generateForVendor but we need to trick it or modify it to only scan "directory" folder.
			// Since generateForVendor scans the WHOLE vendor dir, it's inefficient and might re-generate templates.
			// Let's create a specialized generateForDirectory or just call generateForVendor but beware of overhead?
			// The user asked to apply the "same mechanism".
			// If we call generateForVendor, it will regenerate EVERYTHING. That might be too much if we only want directories.
			// But generateForVendor filters by extension .tpl.
			// Let's implement a targeted generation for "directory" folder.

			targetDir := filepath.Join(outputDir, domainName)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory %s: %w", targetDir, err)
			}

			err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				if filepath.Ext(path) != ".tpl" {
					return nil
				}

				// Load and Execute Template (Logic duplicated from generateForVendor, could be refactored)
				tpl, err := pongo2.FromFile(path)
				if err != nil {
					return fmt.Errorf("failed to load template %s: %w", path, err)
				}

				ctx := pongo2.Context{}
				for k, v := range domainConfig.Variables {
					ctx[k] = v
				}
				ctx["domain_name"] = domainName
				ctx["vendor_name"] = vendor.Name
				ctx["phones"] = domainPhones
				ctx["all_domains"] = allDomains

				out, err := tpl.Execute(ctx)
				if err != nil {
					return fmt.Errorf("failed to render template %s: %w", path, err)
				}

				relPath, err := filepath.Rel(vendor.Dir, path) // Rel to vendor root to keep structure if needed
				if err != nil {
					return err
				}
				// Strip directory/ prefix
				if strings.HasPrefix(relPath, "directory"+string(os.PathSeparator)) {
					relPath = strings.TrimPrefix(relPath, "directory"+string(os.PathSeparator))
				}

				outputRelPath := strings.TrimSuffix(relPath, ".tpl")
				targetFile := filepath.Join(targetDir, outputRelPath)

				if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
					return fmt.Errorf("failed to create directory for file %s: %w", targetFile, err)
				}

				if err := os.WriteFile(targetFile, []byte(out), 0644); err != nil {
					return fmt.Errorf("failed to write config file %s: %w", targetFile, err)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to walk directory folder for vendor %s: %w", vendor.Name, err)
			}
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
func (m *Manager) GeneratePhoneConfigs(outputDir string, phones []models.Phone) ([]string, error) {
	// ... (existing implementation)
	var warnings []string

	for _, phone := range phones {
		// ... (existing loop body)
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
			msg := fmt.Sprintf("Warning: Vendor %s not found for phone %s", phone.Vendor, mac)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}

		if vendor.PhoneConfigFile == "" || vendor.PhoneConfigTemplate == "" {
			continue // Вендор не поддерживает генерацию конфигов телефонов
		}

		// Process Lines and Keys
		var lines []map[string]interface{}
		var keysConfig []string

		// Map features for quick lookup
		featuresMap := make(map[string]Feature)
		for _, f := range vendor.Features {
			featuresMap[f.ID] = f
		}

		// Calculate Max Lines and Used Lines
		maxLines := 1 // Default
		if model, ok := m.getModelByID(phone.ModelID); ok {
			maxLines = model.MaxAccountLines
		}

		linesRange := make([]int, maxLines)
		for i := 0; i < maxLines; i++ {
			linesRange[i] = i + 1
		}

		usedLineNumbers := make(map[int]bool)
		for _, l := range phone.Lines {
			if l.Type == "line" {
				usedLineNumbers[l.Number] = true
			}
		}

		for _, l := range phone.Lines {
			lineData := map[string]interface{}{
				"type":                    l.Type,
				"number":                  l.Number,
				"expansion_module_number": l.ExpansionModuleNumber,
				"key_number":              l.KeyNumber,
			}

			// Parse AdditionalInfo
			var additionalInfo map[string]interface{}
			if l.AdditionalInfo != "" {
				if err := json.Unmarshal([]byte(l.AdditionalInfo), &additionalInfo); err == nil {
					for k, v := range additionalInfo {
						lineData[k] = v
					}
				}
			}
			lines = append(lines, lineData)

			// Generate Key Config if it matches a feature
			featureType := l.Type
			if t, ok := lineData["type"].(string); ok && t != "" {
				featureType = t
			}

			if feature, ok := featuresMap[featureType]; ok {
				for _, param := range feature.Params {
					var val interface{}
					if param.Value != "" {
						val = param.Value
					} else {
						if v, exists := lineData[param.ID]; exists {
							val = v
						}
					}

					if val != nil && param.ConfigTemplate != "" {
						tmpl := param.ConfigTemplate
						tmpl = strings.ReplaceAll(tmpl, "{{key_index}}", fmt.Sprintf("%d", l.Number))
						tmpl = strings.ReplaceAll(tmpl, "{{value}}", fmt.Sprintf("%v", val))
						keysConfig = append(keysConfig, tmpl)
					}
				}
			}
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
				"id":                phone.ID,
				"domain":            phone.Domain,
				"vendor":            phone.Vendor,
				"mac_address":       mac,
				"phone_number":      number,
				"ip_address":        phone.IPAddress,
				"type":              phone.Type,
				"lines":             lines,
				"max_lines":         maxLines,
				"lines_range":       linesRange,
				"used_line_numbers": usedLineNumbers,
			},
			"domain":      domainCtx,
			"keys_config": keysConfig,
		}

		// 1. Генерируем имя файла
		filenameTpl, err := pongo2.FromString(vendor.PhoneConfigFile)
		if err != nil {
			msg := fmt.Sprintf("Warning: Failed to parse filename template for vendor %s: %v", vendor.Name, err)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}
		filename, err := filenameTpl.Execute(ctx)
		if err != nil {
			msg := fmt.Sprintf("Warning: Failed to execute filename template for phone %s: %v", mac, err)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}

		// 2. Генерируем содержимое
		templatePath := filepath.Join(vendor.Dir, vendor.PhoneConfigTemplate)
		tpl, err := pongo2.FromFile(templatePath)
		if err != nil {
			msg := fmt.Sprintf("Warning: Failed to load template %s: %v", templatePath, err)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}
		content, err := tpl.Execute(ctx)
		if err != nil {
			msg := fmt.Sprintf("Warning: Failed to execute template for phone %s: %v", mac, err)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}

		// 3. Записываем файл
		targetDir := filepath.Join(outputDir, phone.Domain)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return warnings, fmt.Errorf("failed to create directory %s: %v", targetDir, err)
		}
		targetFile := filepath.Join(targetDir, filename)
		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			msg := fmt.Sprintf("Warning: Failed to write config file %s: %v", targetFile, err)
			log.Println(msg)
			warnings = append(warnings, msg)
			continue
		}
	}
	return warnings, nil
}

// DeletePhoneConfig удаляет конфигурационный файл для телефона
func (m *Manager) DeletePhoneConfig(outputDir string, phone models.Phone) error {
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
		// Если вендор не найден, мы не можем узнать имя файла, чтобы удалить его.
		// Это не обязательно ошибка, может конфига и не было.
		return fmt.Errorf("vendor %s not found", phone.Vendor)
	}

	if vendor.PhoneConfigFile == "" {
		return nil // Нет шаблона имени файла -> нет файла
	}

	// Подготавливаем контекст (минимальный, нужен только для имени файла)
	// Для имени файла обычно нужен mac или number.
	// Lines и прочее не нужны для имени файла, но на всякий случай передадим пустые.

	ctx := pongo2.Context{
		"account": map[string]interface{}{
			"id":           phone.ID,
			"domain":       phone.Domain,
			"vendor":       phone.Vendor,
			"mac_address":  mac,
			"phone_number": number,
			"ip_address":   phone.IPAddress,
			"type":         phone.Type,
		},
		// domain context might be needed if filename depends on domain variables (unlikely but possible)
	}

	// Генерируем имя файла
	filenameTpl, err := pongo2.FromString(vendor.PhoneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse filename template: %w", err)
	}
	filename, err := filenameTpl.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute filename template: %w", err)
	}

	targetFile := filepath.Join(outputDir, phone.Domain, filename)

	// Удаляем файл
	if err := os.Remove(targetFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file %s: %w", targetFile, err)
	}

	return nil
}

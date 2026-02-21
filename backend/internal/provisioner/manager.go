package provisioner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"provisioning-system/internal/config"
	"provisioning-system/internal/logger"
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
					logger.Warn("Failed to read features file %s: %v", featuresPath, err)
				} else {
					var features []Feature
					if err := yaml.Unmarshal(featuresData, &features); err != nil {
						logger.Warn("Failed to parse features file %s: %v", featuresPath, err)
					} else {
						vc.Features = features
					}
				}
			}
		}

		// Load Accounts if defined
		if vc.AccountsFile != "" {
			accountsPath := filepath.Join(vendorDir, vc.AccountsFile)
			if _, err := os.Stat(accountsPath); err == nil {
				accountsData, err := os.ReadFile(accountsPath)
				if err != nil {
					logger.Warn("Failed to read accounts file %s: %v", accountsPath, err)
				} else {
					var accounts []Feature
					if err := yaml.Unmarshal(accountsData, &accounts); err != nil {
						logger.Warn("Failed to parse accounts file %s: %v", accountsPath, err)
					} else {
						vc.Accounts = accounts
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
		vars := make(map[string]interface{})
		for k, v := range domain.Variables {
			ctx[k] = v
			vars[k] = v
		}
		ctx["variables"] = vars
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

func (m *Manager) GeneratePhoneConfigs(outputDir string, phones []models.Phone) ([]string, error) {
	var warnings []string

	for _, phone := range phones {
		mac := ""
		if phone.MacAddress != nil {
			mac = strings.ReplaceAll(*phone.MacAddress, ":", "")
		}
		// Log generation start
		logger.Info("Generating config for phone %s (Vendor: %s, Model: %s, Domain: %s)", mac, phone.Vendor, phone.ModelID, phone.Domain)

		vendor, ok := m.getVendorByID(phone.Vendor)
		if !ok || vendor.PhoneConfigFile == "" || vendor.PhoneConfigTemplate == "" {
			logger.Warn("Skip phone %s: vendor %s not found or no config support", mac, phone.Vendor)
			continue // Skip if no config support
		}

		// Prepare Maps for Quick Lookup
		featuresMap := make(map[string]Feature)
		for _, f := range vendor.Features {
			featuresMap[f.ID] = f
		}

		// Map Account Data from DB
		var linesForContext []map[string]interface{}
		phoneAccounts := make(map[int]map[string]interface{})
		for _, l := range phone.Lines {
			if l.Type == "Line" {
				accData := l.GetAdditionalInfoMap()
				accData["account_number"] = l.AccountNumber
				accData["number"] = l.AccountNumber
				accData["type"] = strings.ToLower(l.Type)
				phoneAccounts[l.AccountNumber] = accData
				linesForContext = append(linesForContext, accData)
			}
		}

		// Map DB Assignments (Panel-Key overrides)
		dbLinesMap := make(map[string]models.PhoneLine)
		var generalFeatures []models.PhoneLine
		for _, l := range phone.Lines {
			if l.PanelNumber != nil && l.KeyNumber != nil {
				key := fmt.Sprintf("%d-%d", *l.PanelNumber, *l.KeyNumber)
				dbLinesMap[key] = l
			} else {
				// No button association -> likely a general feature
				generalFeatures = append(generalFeatures, l)
			}
		}

		// Main Rendering Loop - Based on Model
		var keysConfig []string
		model, modelOk := m.getModelByID(phone.ModelID)
		if !modelOk {
			logger.Warn("Skip phone %s: model %s not found", mac, phone.ModelID)
			continue
		}

		// 1. Process Main Device Keys
		for _, mk := range model.Keys {
			keyID := fmt.Sprintf("0-%d", mk.Index)
			dbLine, hasOverride := dbLinesMap[keyID]

			assignmentType := mk.Type
			accNum := mk.Account
			lineData := make(map[string]interface{})

			if hasOverride {
				assignmentType = dbLine.Type
				accNum = dbLine.AccountNumber
				lineData = dbLine.GetAdditionalInfoMap()
			}

			// Base Context
			ctx := pongo2.Context{
				"key_index":        mk.Index,
				"key_number":       mk.Index,
				"panel_number":     0,
				"expansion_module": 0,
				"key_type":         mk.Type,
				"type":             assignmentType,
				"label":            mk.Label,
				"settings":         mk.Settings,
				"x":                mk.X,
				"y":                mk.Y,
			}

			// Add Account Data
			if acc, ok := phoneAccounts[accNum]; ok {
				ctx["account"] = acc
				ctx["account_number"] = accNum
				for k, v := range acc {
					ctx["account_"+k] = v
				}
			}

			// Add Assignment Overrides
			for k, v := range lineData {
				ctx[k] = v
			}

			// Render
			if feature, ok := featuresMap[assignmentType]; ok {
				ctx["feature_name"] = feature.Name
				ctx["name"] = feature.Name
				for _, param := range feature.Params {
					m.renderAndAppend(&keysConfig, param, ctx, mk.Settings, fmt.Sprintf("Key 0-%d (Feature %s)", mk.Index, assignmentType))
				}
			}
		}

		// 2. Process General Features (RingTone, etc.)
		for _, gf := range generalFeatures {
			if feature, ok := featuresMap[gf.Type]; ok {
				gfData := gf.GetAdditionalInfoMap()
				ctx := pongo2.Context{
					"type": gf.Type,
				}
				// Add assignment data
				for k, v := range gfData {
					ctx[k] = v
				}
				// If associated with account, add account context
				if feature.AssociatedWithAccount {
					accNum := gf.AccountNumber
					if acc, ok := phoneAccounts[accNum]; ok {
						ctx["account"] = acc
						ctx["account_number"] = accNum
						for k, v := range acc {
							ctx["account_"+k] = v
						}
					}
				}

				ctx["feature_name"] = feature.Name
				ctx["name"] = feature.Name

				for _, param := range feature.Params {
					m.renderAndAppend(&keysConfig, param, ctx, nil, fmt.Sprintf("General Feature %s", gf.Type))
				}
			}
		}

		// 3. Process Expansion Modules (To be implemented when we have M685/M680 models)
		// ...

		// 3. Render Final Config
		domainConfig := m.Config.GetEffectiveDomainConfig(phone.Domain)
		domainCtx := make(map[string]interface{})
		for k, v := range domainConfig.Variables {
			domainCtx[k] = v
		}

		number := ""
		if phone.PhoneNumber != nil {
			number = *phone.PhoneNumber
		}

		context := pongo2.Context{
			"phone":       phone,
			"vendor":      vendor,
			"variables":   domainCtx,
			"keys_config": keysConfig,
			"account": map[string]interface{}{
				"id":           phone.ID,
				"domain":       phone.Domain,
				"vendor":       phone.Vendor,
				"mac_address":  mac,
				"phone_number": number,
				"ip_address":   phone.IPAddress,
				"type":         phone.Type,
				"lines":        linesForContext,
			},
		}

		// Render main template
		tplPath := filepath.Join(vendor.Dir, vendor.PhoneConfigTemplate)
		tplData, err := os.ReadFile(tplPath)
		if err != nil {
			logger.Error("Error reading phone template %s: %v", tplPath, err)
			continue
		}

		mainTpl, err := pongo2.FromString(string(tplData))
		if err != nil {
			logger.Error("Error parsing phone template %s: %v", tplPath, err)
			continue
		}

		finalConfig, err := mainTpl.Execute(context)
		if err != nil {
			logger.Error("Error executing phone template %s: %v", tplPath, err)
			continue
		}

		// Save to file
		fileNameTpl, err := pongo2.FromString(vendor.PhoneConfigFile)
		if err != nil {
			logger.Error("Error parsing phone config name template: %v", err)
			continue
		}
		fileName, err := fileNameTpl.Execute(context)
		if err != nil {
			logger.Error("Error executing phone config name template: %v", err)
			continue
		}

		fullPath := filepath.Join(outputDir, phone.Domain, fileName)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			logger.Error("Error creating directory %s: %v", filepath.Dir(fullPath), err)
			continue
		}

		if err := os.WriteFile(fullPath, []byte(finalConfig), 0644); err != nil {
			logger.Error("Error writing config file %s: %v", fullPath, err)
			continue
		}
		logger.Info("Config saved to: %s", fullPath)
	}

	return warnings, nil
}

func (m *Manager) renderAndAppend(keysConfig *[]string, param FeatureParam, ctx pongo2.Context, modelSettings map[string]string, debugCtx string) {
	val := ctx[param.ID]
	if val == nil {
		if param.Value != "" {
			val = param.Value
		}
	}

	if val != nil && param.ConfigTemplate != "" {
		renderCtx := make(pongo2.Context)
		for k, v := range ctx {
			renderCtx[k] = v
		}
		renderCtx["value"] = val
		renderCtx["id"] = param.ID

		for k, v := range param.Extra {
			renderCtx[k] = v
		}
		if tag, ok := modelSettings[param.ID]; ok {
			renderCtx["tag"] = tag
		}

		if out, err := renderPongoTemplate(param.ConfigTemplate, renderCtx); err == nil {
			logger.Info("  [%s] Rendering param %s -> %s", debugCtx, param.ID, out)
			*keysConfig = append(*keysConfig, out)
		} else {
			logger.Error("  [%s] Failed to render param %s: %v", debugCtx, param.ID, err)
		}
	}
}

// GetPhoneConfigPath returns the path to the phone's configuration file (without checking its existence)
func (m *Manager) GetPhoneConfigPath(outputDir string, phone models.Phone) (string, error) {
	mac := ""
	if phone.MacAddress != nil {
		mac = strings.ReplaceAll(*phone.MacAddress, ":", "")
	}
	number := ""
	if phone.PhoneNumber != nil {
		number = *phone.PhoneNumber
	}

	vendor, ok := m.getVendorByID(phone.Vendor)
	if !ok {
		return "", fmt.Errorf("vendor %s not found", phone.Vendor)
	}

	if vendor.PhoneConfigFile == "" {
		return "", nil // No filename template -> no file
	}

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
	}

	filenameTpl, err := pongo2.FromString(vendor.PhoneConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse filename template: %w", err)
	}
	filename, err := filenameTpl.Execute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute filename template: %w", err)
	}

	return filepath.Join(outputDir, phone.Domain, filename), nil
}

// DeletePhoneConfig deletes the configuration file for the phone
func (m *Manager) DeletePhoneConfig(outputDir string, phone models.Phone) error {
	path, err := m.GetPhoneConfigPath(outputDir, phone)
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}

	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		logger.Info("Deleting old config file: %s", path)
		return os.Remove(path)
	}
	return nil
}

func renderPongoTemplate(tplString string, ctx pongo2.Context) (string, error) {
	tpl, err := pongo2.FromString(tplString)
	if err != nil {
		return "", err
	}
	return tpl.Execute(ctx)
}

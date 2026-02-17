package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"

	"provisioning-system/internal/logger"
	"provisioning-system/internal/models"
	"provisioning-system/internal/provisioner"

	"gorm.io/gorm"
)

type PhoneHandler struct {
	ConfigDir   string
	DB          *gorm.DB
	ProvManager *provisioner.Manager
}

func NewPhoneHandler(configDir string, db *gorm.DB, pm *provisioner.Manager) *PhoneHandler {
	return &PhoneHandler{
		ConfigDir:   configDir,
		DB:          db,
		ProvManager: pm,
	}
}

// CreatePhone handles POST /api/phones
func (h *PhoneHandler) CreatePhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var phone models.Phone
	if err := json.NewDecoder(r.Body).Decode(&phone); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find model in manager
	var model *provisioner.DeviceModel
	if phone.ModelID != "" {
		for _, m := range h.ProvManager.Models {
			if m.ID == phone.ModelID {
				model = &m
				break
			}
		}
	}

	// Validation Logic
	isGateway := model != nil && model.Type == "gateway"
	phone.Type = "phone"
	if isGateway {
		phone.Type = "gateway"
	}

	if !isGateway {
		if phone.MacAddress == nil || *phone.MacAddress == "" {
			http.Error(w, "MAC Address is required for phones", http.StatusBadRequest)
			return
		}
	} else {
		// Gateway Logic
		if phone.IPAddress == "" {
			http.Error(w, "IP Address is required for gateways", http.StatusBadRequest)
			return
		}
		// Copy IP to PhoneNumber for search
		ip := phone.IPAddress
		phone.PhoneNumber = &ip

		// For gateways, if MAC is empty string, set to nil
		if phone.MacAddress != nil && *phone.MacAddress == "" {
			phone.MacAddress = nil
		}
	}

	if model != nil {
		// Calculate limits
		maxAccountLines := model.MaxAccountLines
		ownSoftKeys := model.OwnSoftKeys
		ownHardKeys := model.OwnHardKeys

		// Expansion modules
		expansionKeys := 0
		if phone.ExpansionModulesCount > 0 && phone.ExpansionModuleModel != "" {
			// Find expansion module model
			var expModel *provisioner.DeviceModel
			for _, m := range h.ProvManager.Models {
				if m.ID == phone.ExpansionModuleModel && m.Type == "expansion-module" {
					expModel = &m
					break
				}
			}
			if expModel != nil {
				expansionKeys = phone.ExpansionModulesCount * expModel.OwnHardKeys
			}
		}

		totalLimit := ownSoftKeys + ownHardKeys + maxAccountLines + expansionKeys
		if isGateway {
			totalLimit = maxAccountLines // For gateway, limit is just max accounts (lines)
		}

		if len(phone.Lines) > totalLimit {
			http.Error(w, fmt.Sprintf("Too many lines. Max allowed: %d", totalLimit), http.StatusBadRequest)
			return
		}

		// Check "Line" type limit and one-account-one-line rule
		usedAccounts := make(map[int]bool)
		lineCount := 0
		for _, l := range phone.Lines {
			if l.Type == "Line" {
				lineCount++
				if usedAccounts[l.AccountNumber] {
					http.Error(w, fmt.Sprintf("Duplicate account %d used for Line type", l.AccountNumber), http.StatusBadRequest)
					return
				}
				usedAccounts[l.AccountNumber] = true
			}
		}
		if lineCount > maxAccountLines {
			http.Error(w, fmt.Sprintf("Too many account lines. Max allowed: %d", maxAccountLines), http.StatusBadRequest)
			return
		}
	}

	// Check for duplicate MAC
	if phone.MacAddress != nil && *phone.MacAddress != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("mac_address = ?", *phone.MacAddress).Count(&count)
		if count > 0 {
			http.Error(w, "Phone with this MAC address already exists", http.StatusConflict)
			return
		}
	}

	// Generate Random Password if enabled
	cfg := h.ProvManager.Config
	domainCfg := cfg.GetEffectiveDomainConfig(phone.Domain)
	if domainCfg.GenerateRandomPassword {
		for i := range phone.Lines {
			if phone.Lines[i].Type == "Line" {
				// Parse AdditionalInfo
				var info map[string]interface{}
				if phone.Lines[i].AdditionalInfo != "" {
					json.Unmarshal([]byte(phone.Lines[i].AdditionalInfo), &info)
				} else {
					info = make(map[string]interface{})
				}

				// Check if password is empty
				if pwd, ok := info["password"].(string); !ok || pwd == "" {
					newPwd := generateRandomPassword(12)
					info["password"] = newPwd

					// Update AdditionalInfo
					if data, err := json.Marshal(info); err == nil {
						phone.Lines[i].AdditionalInfo = string(data)
					}
				}
			}
		}
	}

	if result := h.DB.Create(&phone); result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Generate config for new phone
	outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
	if _, err := h.ProvManager.GeneratePhoneConfigs(outputDir, []models.Phone{phone}); err != nil {
		// Rollback: Delete the phone we just created
		h.DB.Delete(&phone)
		http.Error(w, fmt.Sprintf("Failed to generate configs: %v", err), http.StatusInternalServerError)
		return
	}

	// Deploy to domain
	if err := h.deployDomain(phone.Domain, &phone); err != nil {
		logger.Warn("Failed to deploy domain %s: %v", phone.Domain, err)
		// We don't fail the request if deployment fails, but we log it.
		// Or should we? User might want to know.
		// Let's keep it as is for now, but maybe add a warning header?
	}

	// Regenerate directories
	go func() {
		var allPhones []models.Phone
		if err := h.DB.Preload("Lines").Find(&allPhones).Error; err != nil {
			fmt.Printf("Failed to fetch phones for directory regeneration: %v\n", err)
			return
		}
		outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
		if err := h.ProvManager.GenerateDirectories(outputDir, allPhones); err != nil {
			fmt.Printf("Failed to regenerate directories: %v\n", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(phone)
}

// GetPhones handles GET /api/phones
func (h *PhoneHandler) GetPhones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := h.DB.Model(&models.Phone{})

	// Filters
	if domain := r.URL.Query().Get("domain"); domain != "" {
		query = query.Where("domain = ?", domain)
	}
	if vendor := r.URL.Query().Get("vendor"); vendor != "" {
		query = query.Where("vendor = ?", vendor)
	}
	if modelID := r.URL.Query().Get("model_id"); modelID != "" {
		query = query.Where("model_id = ?", modelID)
	}
	if mac := r.URL.Query().Get("mac"); mac != "" {
		mac = strings.ReplaceAll(mac, "*", "%")
		query = query.Where("mac_address LIKE ?", mac)
	}
	if number := r.URL.Query().Get("number"); number != "" {
		number = strings.ReplaceAll(number, "*", "%")
		query = query.Where("phone_number LIKE ?", number)
	}
	if q := r.URL.Query().Get("q"); q != "" {
		q = "%" + strings.ReplaceAll(q, "*", "%") + "%"
		query = query.Where(
			h.DB.Where("phone_number LIKE ?", q).
				Or("description LIKE ?", q).
				Or("id IN (SELECT phone_id FROM phone_lines WHERE additional_info LIKE ?)", q),
		)
	}

	// Count total
	var total int64
	h.DB.Model(&models.Phone{}).Where(query).Count(&total)

	// Pagination
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := (page - 1) * limit

	var phones []models.Phone
	if result := query.Limit(limit).Offset(offset).Order("id desc").Preload("Lines").Find(&phones); result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch phones: %v", result.Error), http.StatusInternalServerError)
		return
	}

	// Enrich phones with display names
	for i := range phones {
		// Model Name
		for _, m := range h.ProvManager.Models {
			if m.ID == phones[i].ModelID {
				phones[i].ModelName = m.Name
				break
			}
		}
		// Vendor Name
		for _, v := range h.ProvManager.Vendors {
			if v.ID == phones[i].Vendor {
				phones[i].VendorName = v.Name
				break
			}
		}
	}

	response := map[string]interface{}{
		"phones": phones,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PhoneHandler) UpdatePhone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reqPhone models.Phone
	if err := json.NewDecoder(r.Body).Decode(&reqPhone); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingPhone models.Phone
	if err := h.DB.Preload("Lines").First(&existingPhone, id).Error; err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	// Find model in manager
	var model *provisioner.DeviceModel
	if reqPhone.ModelID != "" {
		for _, m := range h.ProvManager.Models {
			if m.ID == reqPhone.ModelID {
				model = &m
				break
			}
		}
	}

	// Validation Logic
	isGateway := model != nil && model.Type == "gateway"
	reqPhone.Type = "phone"
	if isGateway {
		reqPhone.Type = "gateway"
	}

	if !isGateway {
		if reqPhone.MacAddress == nil || *reqPhone.MacAddress == "" {
			http.Error(w, "MAC Address is required for phones", http.StatusBadRequest)
			return
		}
	} else {
		// Gateway Logic
		if reqPhone.IPAddress == "" {
			http.Error(w, "IP Address is required for gateways", http.StatusBadRequest)
			return
		}
		// Copy IP to PhoneNumber for search
		ip := reqPhone.IPAddress
		reqPhone.PhoneNumber = &ip

		if reqPhone.MacAddress != nil && *reqPhone.MacAddress == "" {
			reqPhone.MacAddress = nil
		}
	}

	if model != nil {
		maxAccountLines := model.MaxAccountLines

		// 1. Check Account Lines
		usedAccounts := make(map[int]bool)
		lineCount := 0
		for _, l := range reqPhone.Lines {
			if l.Type == "Line" {
				lineCount++
				if usedAccounts[l.AccountNumber] {
					http.Error(w, fmt.Sprintf("Duplicate account %d used for Line type", l.AccountNumber), http.StatusBadRequest)
					return
				}
				usedAccounts[l.AccountNumber] = true
			}
		}
		if lineCount > maxAccountLines {
			http.Error(w, fmt.Sprintf("Too many account lines. Max allowed: %d", maxAccountLines), http.StatusBadRequest)
			return
		}

		// 2. Check Key Limits (Main Phone)
		// ...

		// 3. Check Expansion Modules
		if reqPhone.ExpansionModulesCount > 0 && reqPhone.ExpansionModuleModel != "" {
			var expModel *provisioner.DeviceModel
			for _, m := range h.ProvManager.Models {
				if m.ID == reqPhone.ExpansionModuleModel && m.Type == "expansion-module" {
					expModel = &m
					break
				}
			}
			if expModel != nil {
				for i := 1; i <= reqPhone.ExpansionModulesCount; i++ {
					count := 0
					for _, l := range reqPhone.Lines {
						if l.PanelNumber != nil && *l.PanelNumber == i {
							count++
						}
					}
					// For expansion modules, we just check if total keys doesn't exceed its capacity?
					// Actually, with the new model-driven approach, we'll validate against module keys too.
					// For now, keep it simple.
				}
			}
		}
	}

	// Check for duplicate MAC (exclude current phone)
	if reqPhone.MacAddress != nil && *reqPhone.MacAddress != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("mac_address = ? AND id != ?", *reqPhone.MacAddress, id).Count(&count)
		if count > 0 {
			http.Error(w, "Phone with this MAC address already exists", http.StatusConflict)
			return
		}
	}

	// Update fields in memory to test config generation
	tempPhone := existingPhone
	tempPhone.Domain = reqPhone.Domain
	tempPhone.Vendor = reqPhone.Vendor
	tempPhone.ModelID = reqPhone.ModelID
	tempPhone.MacAddress = reqPhone.MacAddress
	tempPhone.PhoneNumber = reqPhone.PhoneNumber
	tempPhone.IPAddress = reqPhone.IPAddress
	tempPhone.Description = reqPhone.Description
	tempPhone.ExpansionModulesCount = reqPhone.ExpansionModulesCount
	tempPhone.ExpansionModuleModel = reqPhone.ExpansionModuleModel
	tempPhone.Type = reqPhone.Type
	tempPhone.Lines = reqPhone.Lines // This is a slice, so it's a reference, but GeneratePhoneConfigs reads it.

	// Determine if config path will change
	outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
	oldPath, _ := h.ProvManager.GetPhoneConfigPath(outputDir, existingPhone)
	newPath, _ := h.ProvManager.GetPhoneConfigPath(outputDir, tempPhone)

	logger.Info("[UpdatePhone] Old config path: %s", oldPath)
	logger.Info("[UpdatePhone] New config path: %s", newPath)

	if oldPath != newPath && oldPath != "" {
		logger.Info("[UpdatePhone] Paths differ, deleting old config: %s", oldPath)
		if err := h.ProvManager.DeletePhoneConfig(outputDir, existingPhone); err != nil {
			logger.Warn("Failed to delete old config %s: %v", oldPath, err)
		}
	} else {
		logger.Info("[UpdatePhone] Paths match, skipping deletion")
	}

	// Generate new config
	logger.Info("[UpdatePhone] Generating new config...")
	if _, err := h.ProvManager.GeneratePhoneConfigs(outputDir, []models.Phone{tempPhone}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate configs: %v", err), http.StatusInternalServerError)
		return
	}
	logger.Info("[UpdatePhone] Generation complete")

	// Apply updates to DB
	existingPhone = tempPhone // Copy fields back (except Lines which need association update)

	// Update Lines using Association Replace
	if err := h.DB.Session(&gorm.Session{FullSaveAssociations: true}).Model(&existingPhone).Association("Lines").Replace(reqPhone.Lines); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update lines: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the phone itself
	if err := h.DB.Save(&existingPhone).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to update phone: %v", err), http.StatusInternalServerError)
		return
	}

	// Deploy to domain
	if err := h.deployDomain(existingPhone.Domain, &existingPhone); err != nil {
		fmt.Printf("Failed to deploy domain %s: %v\n", existingPhone.Domain, err)
	}

	// Regenerate directories
	go func() {
		var allPhones []models.Phone
		if err := h.DB.Preload("Lines").Find(&allPhones).Error; err != nil {
			fmt.Printf("Failed to fetch phones for directory regeneration: %v\n", err)
			return
		}
		outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
		if err := h.ProvManager.GenerateDirectories(outputDir, allPhones); err != nil {
			fmt.Printf("Failed to regenerate directories: %v\n", err)
		}
	}()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingPhone)
}

func (h *PhoneHandler) deployDomain(domainName string, phone *models.Phone) error {
	cfg := h.ProvManager.Config
	domainCfg := cfg.GetEffectiveDomainConfig(domainName)

	if len(domainCfg.DeployCommands) == 0 {
		return nil // No deploy commands, nothing to do
	}

	return h.executeCommands(domainCfg.DeployCommands, domainName, phone, domainCfg.Variables)
}

// GetVendors handles GET /api/vendors
func (h *PhoneHandler) GetVendors(w http.ResponseWriter, r *http.Request) {
	var vendors []map[string]interface{}
	for _, v := range h.ProvManager.Vendors {
		vendors = append(vendors, map[string]interface{}{
			"id":       v.ID,
			"name":     v.Name,
			"features": v.Features,
			"accounts": v.Accounts,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vendors": vendors,
	})
}

// GetModels handles GET /api/models
// Query params: vendor
func (h *PhoneHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	vendor := r.URL.Query().Get("vendor")
	var models []provisioner.DeviceModel

	for _, m := range h.ProvManager.Models {
		if m.Type == "expansion-module" {
			continue
		}
		if vendor == "" || strings.EqualFold(m.Vendor, vendor) {
			models = append(models, m)
		}
	}

	// Sort models by Name
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// DeletePhone handles DELETE /api/phones/{id}
func (h *PhoneHandler) DeletePhone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var phone models.Phone
	if err := h.DB.Preload("Lines").First(&phone, id).Error; err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	// 1. Delete config file
	outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
	if err := h.ProvManager.DeletePhoneConfig(outputDir, phone); err != nil {
		logger.Warn("Failed to delete config file: %v", err)
	}

	// 2. Delete from DB
	// Delete associated lines first (GORM should handle this with cascading or we do it manually)
	// Using Select("Lines").Delete(&phone) deletes the phone and clears associations, but might not delete Line records if not configured.
	// Let's rely on GORM's association handling or manual cleanup if needed.
	// For now, simple Delete.
	if err := h.DB.Select("Lines").Delete(&phone).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete phone: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Execute DeleteCmd (Deploy changes)
	if err := h.executeDeleteCmd(phone.Domain, &phone); err != nil {
		logger.Warn("Failed to execute delete command for domain %s: %v", phone.Domain, err)
		// We don't fail the request if the hook fails, but we log it.
	}

	// Regenerate directories
	go func() {
		var allPhones []models.Phone
		if err := h.DB.Preload("Lines").Find(&allPhones).Error; err != nil {
			logger.Error("Failed to fetch phones for directory regeneration: %v", err)
			return
		}
		outputDir := strings.TrimSuffix(h.ConfigDir, "/") + "/temp_configs"
		if err := h.ProvManager.GenerateDirectories(outputDir, allPhones); err != nil {
			fmt.Printf("Failed to regenerate directories: %v\n", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Phone deleted successfully"}`))
}

func (h *PhoneHandler) executeDeleteCmd(domainName string, phone *models.Phone) error {
	cfg := h.ProvManager.Config
	domainCfg := cfg.GetEffectiveDomainConfig(domainName)

	if len(domainCfg.DeleteCommands) == 0 {
		return nil
	}

	return h.executeCommands(domainCfg.DeleteCommands, domainName, phone, domainCfg.Variables)
}

func (h *PhoneHandler) executeCommands(commands []string, domainName string, phone *models.Phone, domainVars map[string]string) error {
	sourceDir, _ := filepath.Abs(filepath.Join(h.ConfigDir, "temp_configs", domainName))

	// Data for template
	data := struct {
		Phone  *models.Phone
		Domain string
		Vars   map[string]string
	}{
		Phone:  phone,
		Domain: domainName,
		Vars:   domainVars,
	}

	for _, cmdStr := range commands {
		// 1. Parse template
		tmpl, err := template.New("cmd").Parse(cmdStr)
		if err != nil {
			return fmt.Errorf("failed to parse command template '%s': %w", cmdStr, err)
		}

		var cmdBuf bytes.Buffer
		if err := tmpl.Execute(&cmdBuf, data); err != nil {
			return fmt.Errorf("failed to execute command template '%s': %w", cmdStr, err)
		}

		finalCmdStr := cmdBuf.String()

		// 2. Execute command
		// We use "sh -c" to allow complex commands (pipes, redirects, etc.)
		cmd := exec.Command("sh", "-c", finalCmdStr)

		// Set environment variables
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PROVISIONING_DOMAIN=%s", domainName),
			fmt.Sprintf("PROVISIONING_SOURCE=%s", sourceDir),
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command execution failed: '%s', error: %v, output: %s", finalCmdStr, err, string(output))
		}
	}

	return nil
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

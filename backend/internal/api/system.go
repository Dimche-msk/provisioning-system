package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"provisioning-system/internal/config"
	"provisioning-system/internal/models"
	"provisioning-system/internal/provisioner"

	"gorm.io/gorm"
)

type SystemHandler struct {
	ConfigDir   string
	Config      **config.SystemConfig // Pointer to the global config pointer to allow updates
	ProvManager *provisioner.Manager
	DB          *gorm.DB
}

func NewSystemHandler(configDir string, cfg **config.SystemConfig, pm *provisioner.Manager, db *gorm.DB) *SystemHandler {
	return &SystemHandler{
		ConfigDir:   configDir,
		Config:      cfg,
		ProvManager: pm,
		DB:          db,
	}
}

func (h *SystemHandler) Reload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Reload provisioning-system.yaml
	newCfg, err := config.LoadConfig(h.ConfigDir)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to reload config: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update global config
	*h.Config = newCfg
	h.ProvManager.Config = newCfg

	// 2. Reload vendor configs
	vendorsDir := filepath.Join(h.ConfigDir, "vendors")
	if err := h.ProvManager.LoadVendors(vendorsDir); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to reload vendors: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// 3. Generate configs
	outputDir := filepath.Join(h.ConfigDir, "temp_configs")
	backupDir := filepath.Join(h.ConfigDir, "temp_configs.bak")

	if _, err := os.Stat(outputDir); err == nil {
		if err := os.RemoveAll(backupDir); err != nil {
			log.Printf("Warning: Failed to remove old backup: %v", err)
		}
		if err := os.Rename(outputDir, backupDir); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to backup configs: %v"}`, err), http.StatusInternalServerError)
			return
		}
	}
	// 4. Генерация основных конфигов
	if err := h.ProvManager.GenerateConfigs(outputDir); err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate configs: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Генерация конфигов телефонов
	var phones []models.Phone
	if result := h.DB.Find(&phones); result.Error != nil {
		log.Printf("Failed to fetch phones for config generation: %v", result.Error)
		// Не прерываем процесс, так как основные конфиги уже сгенерированы
	} else {
		if err := h.ProvManager.GeneratePhoneConfigs(outputDir, phones); err != nil {
			log.Printf("Failed to generate phone configs: %v", err)
			// Тоже не прерываем, но логируем
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Configuration reloaded and generated successfully"}`))
}

func (h *SystemHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
	cfg := *h.Config

	// Collect unique domains (Defaults + specific Domains)
	// Using a map to ensure uniqueness if needed, though the slice should be enough if config is valid
	domains := []string{}

	// Always add default if it has a name, or just "default"
	defaultName := cfg.Defaults.Name
	if defaultName == "" {
		defaultName = "default"
	}
	domains = append(domains, defaultName)

	for _, d := range cfg.Domains {
		if d.Name != defaultName {
			domains = append(domains, d.Name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domains": domains,
	})
}

type DeployRequest struct {
	Domain string `json:"domain"`
}

func (h *SystemHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	domainName := req.Domain
	if domainName == "" {
		http.Error(w, "Domain name is required", http.StatusBadRequest)
		return
	}

	// Get effective config for the domain
	cfg := *h.Config
	domainCfg := cfg.GetEffectiveDomainConfig(domainName)

	if domainCfg.DeployCmd == "" {
		http.Error(w, fmt.Sprintf(`{"error": "No deploy command defined for domain '%s'"}`, domainName), http.StatusBadRequest)
		return
	}

	// Execute command
	// We pass the temp_configs/<domain> path as an environment variable PROVISIONING_SOURCE
	// and also as a simple replacement {{ source }} if we want to support simple templating later.
	// For now, just executing the command.

	cmdParts := strings.Fields(domainCfg.DeployCmd)
	if len(cmdParts) == 0 {
		http.Error(w, "Empty deploy command", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)

	// Set environment variables
	sourceDir, _ := filepath.Abs(filepath.Join(h.ConfigDir, "temp_configs", domainName))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PROVISIONING_DOMAIN=%s", domainName),
		fmt.Sprintf("PROVISIONING_SOURCE=%s", sourceDir),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Deploy failed for %s: %v. Output: %s", domainName, err, string(output))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  fmt.Sprintf("Deploy failed: %v", err),
			"output": string(output),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": fmt.Sprintf("Deployed successfully to %s", domainName),
		"output":  string(output),
	})
}

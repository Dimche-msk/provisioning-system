package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"provisioning-system/internal/backup"
	"provisioning-system/internal/config"
	"provisioning-system/internal/db"
	"provisioning-system/internal/logger"
	"provisioning-system/internal/models"
	"provisioning-system/internal/provisioner"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SystemHandler struct {
	ConfigDir     string
	Config        **config.SystemConfig
	ProvManager   *provisioner.Manager
	DB            *gorm.DB
	BackupManager *backup.Manager
}

func NewSystemHandler(configDir string, cfg **config.SystemConfig, pm *provisioner.Manager, db *gorm.DB, bm *backup.Manager) *SystemHandler {
	return &SystemHandler{
		ConfigDir:     configDir,
		Config:        cfg,
		ProvManager:   pm,
		DB:            db,
		BackupManager: bm,
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to reload config: %v", err)})
		return
	}

	// Update global config
	*h.Config = newCfg
	h.ProvManager.Config = newCfg

	// 2. Reload vendor configs
	vendorsDir := filepath.Join(h.ConfigDir, "vendors")
	if err := h.ProvManager.LoadVendors(vendorsDir); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to reload vendors: %v", err)})
		return
	}
	if err := h.ProvManager.LoadModels(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to reload models: %v", err)})
		return
	}

	// 3. Generate configs
	// 3. Generate configs
	outputDir := filepath.Join(h.ConfigDir, "pre_configs")

	// Clean pre_configs if exists
	if err := os.RemoveAll(outputDir); err != nil {
		log.Printf("Warning: Failed to clean pre_configs: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to create pre_configs dir: %v", err)})
		return
	}

	// 4. Fetch phones (needed for both general configs (directory) and phone configs)
	var phones []models.Phone
	var warnings []string

	if result := h.DB.Preload("Lines").Find(&phones); result.Error != nil {
		log.Printf("Failed to fetch phones for config generation: %v", result.Error)
		warnings = append(warnings, fmt.Sprintf("Failed to fetch phones: %v", result.Error))
	}

	// 5. Generate general configs (including directories)
	if err := h.ProvManager.GenerateConfigs(outputDir, phones); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to generate configs: %v", err)})
		return
	}

	// 6. Generate phone configs
	if len(phones) > 0 {
		phoneWarnings, err := h.ProvManager.GeneratePhoneConfigs(outputDir, phones)
		if err != nil {
			log.Printf("Failed to generate phone configs: %v", err)
			warnings = append(warnings, fmt.Sprintf("Failed to generate phone configs: %v", err))
		}
		if len(phoneWarnings) > 0 {
			warnings = append(warnings, phoneWarnings...)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"message":  "Configuration prepared successfully. Ready to apply.",
		"warnings": warnings,
	})
}

func (h *SystemHandler) ApplyConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	preDir := filepath.Join(h.ConfigDir, "pre_configs")
	targetDir := filepath.Join(h.ConfigDir, "temp_configs")
	backupDir := filepath.Join(h.ConfigDir, "temp_configs.bak")

	// Check if pre_configs exists
	if _, err := os.Stat(preDir); os.IsNotExist(err) {
		http.Error(w, `{"error": "No prepared configuration found. Please prepare configuration first."}`, http.StatusBadRequest)
		return
	}

	// Backup existing temp_configs
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(backupDir); err != nil {
			log.Printf("Warning: Failed to remove old backup: %v", err)
		}
		if err := os.Rename(targetDir, backupDir); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Failed to backup current configs: %v"}`, err), http.StatusInternalServerError)
			return
		}
	}

	// Move pre_configs to temp_configs
	if err := os.Rename(preDir, targetDir); err != nil {
		// Try to restore backup if move failed
		os.Rename(backupDir, targetDir)
		http.Error(w, fmt.Sprintf(`{"error": "Failed to apply configuration: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Configuration applied successfully."}`))
}

func (h *SystemHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
	cfg := *h.Config

	// Collect unique domains (Defaults + specific Domains)
	// Using a map to ensure uniqueness if needed, though the slice should be enough if config is valid
	domains := []string{}

	// Collect detailed domains for frontend usage (e.g. variables)
	detailedDomains := []config.DomainSettings{}
	for _, d := range cfg.Domains {
		domains = append(domains, d.Name)
		detailedDomains = append(detailedDomains, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domains":          domains,
		"detailed_domains": detailedDomains,
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

func (h *SystemHandler) CreateDBBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.BackupManager.CreateBackup(backup.BackupTypeDB); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create DB backup: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Database backup created successfully"}`))
}

func (h *SystemHandler) CreateConfigBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.BackupManager.CreateBackup(backup.BackupTypeConfig); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create Config backup: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Configuration backup created successfully"}`))
}

func (h *SystemHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	backups, err := h.BackupManager.ListBackups()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to list backups: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
	})
}

func (h *SystemHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Basic security check to prevent traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join((*h.Config).Database.BackupDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, filePath)
}

func (h *SystemHandler) UploadBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 10MB limit
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to get file from request: %v"}`, err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".zip") {
		http.Error(w, `{"error": "Only .zip files are allowed"}`, http.StatusBadRequest)
		return
	}

	targetPath := filepath.Join((*h.Config).Database.BackupDir, header.Filename)
	out, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create local file: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to save file: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Backup uploaded successfully"}`))
}

func (h *SystemHandler) RestoreDBBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	if err := h.BackupManager.RestoreDB(req.Filename); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to restore database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Re-initialize Database connection
	newDB, err := db.Init((*h.Config).Database.Path)
	if err != nil {
		log.Printf("CRITICAL: Failed to re-initialize database after restore: %v", err)
		http.Error(w, fmt.Sprintf(`{"error": "Restore successful but DB re-init failed: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Update the global DB instance
	*h.DB = *newDB

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Database restored successfully"}`))
}

func (h *SystemHandler) RestoreConfigBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode restore request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	logger.Info("RestoreConfigBackup request received for: %s", req.Filename)

	if err := h.BackupManager.RestoreConfig(req.Filename); err != nil {
		logger.Error("RestoreConfig failed for %s: %v", req.Filename, err)
		http.Error(w, fmt.Sprintf(`{"error": "Failed to restore configuration: %v"}`, err), http.StatusInternalServerError)
		return
	}

	logger.Info("RestoreConfig successfully completed for: %s", req.Filename)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "message": "Configuration restored successfully. Please 'Reload' and 'Apply' to finalize."}`))
}

type DashboardStats struct {
	Domain string `json:"domain"`
	Vendor string `json:"vendor"`
	Count  int64  `json:"count"`
}

func (h *SystemHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	var stats []DashboardStats
	// GORM query to group by domain and vendor
	if err := h.DB.Model(&models.Phone{}).Select("domain, vendor, count(*) as count").Group("domain, vendor").Scan(&stats).Error; err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to fetch stats: %v"}`, err), http.StatusInternalServerError)
		return
	}

	var totalPhones int64
	h.DB.Model(&models.Phone{}).Count(&totalPhones)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":        stats,
		"total_phones": totalPhones,
	})
}

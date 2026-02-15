package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"provisioning-system/internal/config"
	"provisioning-system/internal/logger"

	"gorm.io/gorm"
)

type Manager struct {
	Config    *config.SystemConfig
	DB        *gorm.DB
	ConfigDir string
}

type BackupType string

const (
	BackupTypeDB      BackupType = "db"
	BackupTypeConfig  BackupType = "config"
	BackupTypeSupport BackupType = "support"
)

type BackupInfo struct {
	Name string     `json:"name"`
	Type BackupType `json:"type"`
	Size int64      `json:"size"`
	Time time.Time  `json:"time"`
}

func NewManager(cfg *config.SystemConfig, db *gorm.DB, configDir string) *Manager {
	return &Manager{
		Config:    cfg,
		DB:        db,
		ConfigDir: configDir,
	}
}

// CreateBackup creates a new backup of the database
func (m *Manager) CreateBackup(bType BackupType) error {
	backupDir := m.Config.Database.BackupDir
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	var zipName string

	if bType == BackupTypeDB {
		// 1. Create temp backup using VACUUM INTO
		tempBackup := filepath.Join(backupDir, "temp_backup.db")
		if err := os.Remove(tempBackup); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old temp backup: %w", err)
		}

		// VACUUM INTO is safe and non-blocking for readers
		if err := m.DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", tempBackup)).Error; err != nil {
			return fmt.Errorf("failed to create db snapshot: %w", err)
		}
		defer os.Remove(tempBackup)

		zipName = fmt.Sprintf("db_%s.zip", timestamp)
		zipPath := filepath.Join(backupDir, zipName)

		if err := zipFile(tempBackup, zipPath, "provisioning.db"); err != nil {
			return fmt.Errorf("failed to compress backup: %w", err)
		}
	} else {
		// Config backup
		zipName = fmt.Sprintf("cfg_%s.zip", timestamp)
		zipPath := filepath.Join(backupDir, zipName)

		if err := zipDir(m.ConfigDir, zipPath); err != nil {
			return fmt.Errorf("failed to compress config backup: %w", err)
		}
	}

	// 3. Rotate backups
	return m.rotateBackups(bType)
}

// ListBackups returns a list of available backups
func (m *Manager) ListBackups() ([]BackupInfo, error) {
	backupDir := m.Config.Database.BackupDir
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
			name := entry.Name()
			var bType BackupType
			if strings.HasPrefix(name, "db_") {
				bType = BackupTypeDB
			} else if strings.HasPrefix(name, "cfg_") {
				bType = BackupTypeConfig
			} else if strings.HasPrefix(name, "support_bundle_") {
				bType = BackupTypeSupport
			} else {
				continue // Ignore other zip files
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, BackupInfo{
				Name: name,
				Type: bType,
				Size: info.Size(),
				Time: info.ModTime(),
			})
		}
	}

	// Sort by time desc
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Time.After(backups[j].Time)
	})

	return backups, nil
}

// RestoreDB restores the database from a backup file
func (m *Manager) RestoreDB(filename string) error {
	backupPath := filepath.Join(m.Config.Database.BackupDir, filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found")
	}

	// 1. Close current DB
	sqlDB, err := m.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	// 2. Unzip and overwrite DB file
	dbPath := m.Config.Database.Path
	if err := unzipFile(backupPath, dbPath); err != nil {
		return fmt.Errorf("failed to restore db file: %w", err)
	}

	return nil
}

// RestoreConfig restores the configuration directory from a backup file
func (m *Manager) RestoreConfig(filename string) error {
	backupPath := filepath.Join(m.Config.Database.BackupDir, filename)
	logger.Info("Starting clean configuration restore from: %s", filename)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		logger.Error("Backup file not found at: %s", backupPath)
		return fmt.Errorf("backup file not found")
	}

	// 1. Clean destination (except backups/db if they are inside)
	absConfigDir, _ := filepath.Abs(m.ConfigDir)
	absBackupDir, _ := filepath.Abs(m.Config.Database.BackupDir)
	absDBPath, _ := filepath.Abs(m.Config.Database.Path)

	logger.Debug("Cleaning ConfigDir: %s", absConfigDir)
	entries, err := os.ReadDir(m.ConfigDir)
	if err == nil {
		for _, entry := range entries {
			entryPath := filepath.Join(m.ConfigDir, entry.Name())
			absEntryPath, _ := filepath.Abs(entryPath)

			// Safety: don't delete if it contains backups or DB
			if isSubpath(absEntryPath, absBackupDir) || isSubpath(absEntryPath, absDBPath) ||
				absEntryPath == absBackupDir || absEntryPath == absDBPath {
				logger.Debug("Skipping protected path during clean: %s", entryPath)
				continue
			}

			logger.Debug("Removing existing item before restore: %s", entryPath)
			os.RemoveAll(entryPath)
		}
	}

	// 2. Overwrite ConfigDir
	if err := unzipDir(backupPath, m.ConfigDir); err != nil {
		logger.Error("Failed to extract configuration: %v", err)
		return fmt.Errorf("failed to restore config: %w", err)
	}

	logger.Info("Configuration extraction to %s completed successfully", m.ConfigDir)
	return nil
}

// DeleteBackup deletes a backup file
func (m *Manager) DeleteBackup(filename string) error {
	// Basic security check to prevent traversal (redundant but safe)
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename")
	}

	backupDir := m.Config.Database.BackupDir
	filePath := filepath.Join(backupDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found")
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	logger.Info("Backup deleted: %s", filename)
	return nil
}

func isSubpath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

func (m *Manager) rotateBackups(bType BackupType) error {
	backups, err := m.ListBackups()
	if err != nil {
		return err
	}

	// Filter by type
	var filtered []BackupInfo
	for _, b := range backups {
		if b.Type == bType {
			filtered = append(filtered, b)
		}
	}

	maxBackups := 5
	if len(filtered) > maxBackups {
		for i := maxBackups; i < len(filtered); i++ {
			path := filepath.Join(m.Config.Database.BackupDir, filtered[i].Name)
			os.Remove(path)
		}
	}
	return nil
}

func (m *Manager) AddFileToZip(zw *zip.Writer, source, nameInZip string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = nameInZip
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

func (m *Manager) AddDirToZip(zw *zip.Writer, source, prefix string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		if prefix != "" {
			header.Name = filepath.Join(prefix, relPath)
		} else {
			header.Name = relPath
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func zipFile(source, target, nameInZip string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = nameInZip
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

func zipDir(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func unzipFile(zipPath, targetPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		return err
	}
	return fmt.Errorf("no files found in zip")
}

func unzipDir(zipPath, targetDir string) error {
	logger.Debug("Unzipping directory: %s to %s", zipPath, targetDir)
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(targetDir, f.Name)
		logger.Debug("Extracting: %s -> %s", f.Name, path)

		if f.FileInfo().IsDir() {
			logger.Debug("Creating directory: %s", path)
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			logger.Error("Failed to create parent directory for %s: %v", path, err)
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			logger.Error("Failed to open output file %s: %v", path, err)
			return err
		}

		rc, err := f.Open()
		if err != nil {
			logger.Error("Failed to open zip entry %s: %v", f.Name, err)
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			logger.Error("Failed to copy data to %s: %v", path, err)
			return err
		}
	}
	return nil
}

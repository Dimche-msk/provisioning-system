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

	"gorm.io/gorm"
)

type Manager struct {
	Config *config.SystemConfig
	DB     *gorm.DB
}

type BackupInfo struct {
	Name string    `json:"name"`
	Size int64     `json:"size"`
	Time time.Time `json:"time"`
}

func NewManager(cfg *config.SystemConfig, db *gorm.DB) *Manager {
	return &Manager{
		Config: cfg,
		DB:     db,
	}
}

// CreateBackup creates a new backup of the database
func (m *Manager) CreateBackup() error {
	backupDir := m.Config.Database.BackupDir
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

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

	// 2. Compress to ZIP
	timestamp := time.Now().Format("20060102_150405")
	zipName := fmt.Sprintf("backup_%s.zip", timestamp)
	zipPath := filepath.Join(backupDir, zipName)

	if err := zipFile(tempBackup, zipPath, "provisioning.db"); err != nil {
		return fmt.Errorf("failed to compress backup: %w", err)
	}

	// 3. Rotate backups
	return m.rotateBackups()
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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") && strings.HasPrefix(entry.Name(), "backup_") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, BackupInfo{
				Name: entry.Name(),
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

// RestoreBackup restores the database from a backup file
// WARNING: This closes the current DB connection. The caller MUST re-initialize the DB.
func (m *Manager) RestoreBackup(filename string) error {
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
	// Backup current DB just in case (optional, but good for safety)
	// os.Rename(dbPath, dbPath+".bak")

	if err := unzipFile(backupPath, dbPath); err != nil {
		return fmt.Errorf("failed to restore db file: %w", err)
	}

	return nil
}

func (m *Manager) rotateBackups() error {
	backups, err := m.ListBackups()
	if err != nil {
		return err
	}

	maxBackups := 5
	if len(backups) > maxBackups {
		for i := maxBackups; i < len(backups); i++ {
			path := filepath.Join(m.Config.Database.BackupDir, backups[i].Name)
			os.Remove(path)
		}
	}
	return nil
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

func unzipFile(zipPath, targetPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// We expect only one file "provisioning.db"
		// But let's just take the first one or match name
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
	return nil
}

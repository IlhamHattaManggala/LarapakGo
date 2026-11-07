package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Disk manages file operations under storage directory.
type Disk struct {
	baseDir string
}

var defaultDisk = &Disk{baseDir: filepath.Join("storage", "app")}

// Default returns the default storage Disk instance.
func Default() *Disk {
	return defaultDisk
}

// Put writes content to the target storage path.
func (d *Disk) Put(path string, content []byte) error {
	fullPath := filepath.Join(d.baseDir, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, 0644)
}

// Get reads content from the target storage path.
func (d *Disk) Get(path string) ([]byte, error) {
	fullPath := filepath.Join(d.baseDir, path)
	return os.ReadFile(fullPath)
}

// Delete removes a file from storage.
func (d *Disk) Delete(path string) error {
	fullPath := filepath.Join(d.baseDir, path)
	return os.Remove(fullPath)
}

// Exists checks if a file exists in storage.
func (d *Disk) Exists(path string) bool {
	fullPath := filepath.Join(d.baseDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// URL returns the public web-accessible URL for a stored asset.
func (d *Disk) URL(path string) string {
	clean := strings.TrimPrefix(filepath.ToSlash(path), "/")
	return "/storage/" + clean
}

// SaveUploaded generates a unique filename, saves the file in the designated subdirectory,
// and returns the relative path suitable for database storage and URL generation.
func (d *Disk) SaveUploaded(content []byte, originalName, subDir string) (string, error) {
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin"
	}

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	randHex := hex.EncodeToString(randBytes)
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s%s", timestamp, randHex, ext)

	relPath := filename
	if subDir != "" {
		relPath = filepath.Join(subDir, filename)
	}

	err := d.Put(relPath, content)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(relPath), nil
}

package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDiskOperations(t *testing.T) {
	tmpDir := t.TempDir()
	disk := &Disk{baseDir: tmpDir}

	// 1. Put
	err := disk.Put("dokumen/cathetan.txt", []byte("Halo saka LarapakGo Storage"))
	if err != nil {
		t.Fatalf("failed to put file: %v", err)
	}

	// 2. Exists
	if !disk.Exists("dokumen/cathetan.txt") {
		t.Fatalf("expected file to exist")
	}

	// 3. Get
	bytes, err := disk.Get("dokumen/cathetan.txt")
	if err != nil || string(bytes) != "Halo saka LarapakGo Storage" {
		t.Fatalf("unexpected content: %s", string(bytes))
	}

	// 4. SaveUploaded
	rel, err := disk.SaveUploaded([]byte("gambar dummy"), "avatar.png", "avatars")
	if err != nil {
		t.Fatalf("failed save uploaded: %v", err)
	}
	if !strings.HasPrefix(rel, "avatars/") || !strings.HasSuffix(rel, ".png") {
		t.Fatalf("unexpected uploaded path format: %s", rel)
	}
	if !disk.Exists(rel) {
		t.Fatalf("uploaded file not found on disk")
	}

	// 5. URL
	url := disk.URL(rel)
	if !strings.HasPrefix(url, "/storage/avatars/") {
		t.Fatalf("unexpected url format: %s", url)
	}

	// 6. Delete
	err = disk.Delete("dokumen/cathetan.txt")
	if err != nil || disk.Exists("dokumen/cathetan.txt") {
		t.Fatalf("failed to delete file")
	}
	_ = filepath.Base(rel)
}

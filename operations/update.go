package operations

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/base-go/cmd/utils"
	"github.com/base-go/cmd/version"
)

func UpdateCore(progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	rawVersion := version.Version
	normalized := strings.TrimPrefix(rawVersion, "v")
	tag := "v" + normalized
	archiveURL := fmt.Sprintf("https://github.com/base-go/base-core/archive/refs/tags/%s.zip", tag)

	tempDir, err := os.MkdirTemp("", "base-core-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	progress(fmt.Sprintf("Downloading core from %s...", archiveURL))
	resp, err := http.Get(archiveURL)
	if err != nil {
		return fmt.Errorf("failed to download core archive: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading core archive", resp.StatusCode)
	}

	tmpZip, err := os.CreateTemp("", "base-core-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %w", err)
	}
	_, err = io.Copy(tmpZip, resp.Body)
	if cerr := tmpZip.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("failed to save core archive: %w", err)
	}
	defer os.Remove(tmpZip.Name())

	progress("Extracting core files...")
	if err := utils.Unzip(tmpZip.Name(), tempDir); err != nil {
		return fmt.Errorf("failed to extract core archive: %w", err)
	}

	candidates := []string{
		filepath.Join(tempDir, fmt.Sprintf("base-%s", tag)),
		filepath.Join(tempDir, fmt.Sprintf("base-%s", normalized)),
	}
	var extractedDir string
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			extractedDir = c
			break
		}
	}
	if extractedDir == "" {
		if entries, err := os.ReadDir(tempDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "base-") {
					extractedDir = filepath.Join(tempDir, e.Name())
					break
				}
			}
		}
	}
	if extractedDir == "" {
		return fmt.Errorf("could not locate extracted base directory for tag %s", tag)
	}

	srcCoreDir := filepath.Join(extractedDir, "core")
	if fi, err := os.Stat(srcCoreDir); err != nil || !fi.IsDir() {
		return fmt.Errorf("core directory not found in archive at %s", srcCoreDir)
	}

	projectCoreDir := filepath.Join(".", "core")
	backupDir := projectCoreDir + ".bak"

	progress("Backing up existing core...")
	if _, err := os.Stat(projectCoreDir); err == nil {
		if err := os.Rename(projectCoreDir, backupDir); err != nil {
			return fmt.Errorf("failed to backup current core directory: %w", err)
		}
	}

	progress("Installing new core files...")
	copyErr := filepath.Walk(srcCoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcCoreDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(projectCoreDir, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, os.ModePerm)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})

	if copyErr != nil {
		os.RemoveAll(projectCoreDir)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.Rename(backupDir, projectCoreDir)
		}
		return fmt.Errorf("failed to copy core files: %w", copyErr)
	}

	os.RemoveAll(backupDir)
	progress("Core updated successfully!")
	return nil
}

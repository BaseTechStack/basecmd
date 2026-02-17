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

func NewProject(projectName string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	rawVersion := version.Version
	normalized := strings.TrimPrefix(rawVersion, "v")
	tag := "v" + normalized
	archiveURL := fmt.Sprintf("https://github.com/base-go/base-core/archive/refs/tags/%s.zip", tag)

	progress("Creating project directory...")
	if err := os.Mkdir(projectName, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	progress("Downloading project template...")
	resp, err := http.Get(archiveURL)
	if err != nil {
		return fmt.Errorf("downloading project template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d when downloading template from %s", resp.StatusCode, archiveURL)
	}

	tmpZip, err := os.CreateTemp("", "base-project-*.zip")
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}
	defer os.Remove(tmpZip.Name())

	bytesWritten, err := io.Copy(tmpZip, resp.Body)
	if err != nil {
		return fmt.Errorf("saving project template: %w", err)
	}
	tmpZip.Close()

	if bytesWritten < 1000 {
		return fmt.Errorf("downloaded file too small (%d bytes), possibly invalid", bytesWritten)
	}

	progress("Extracting project files...")
	if err := utils.Unzip(tmpZip.Name(), projectName); err != nil {
		return fmt.Errorf("extracting project template: %w", err)
	}

	// Locate the extracted subdirectory
	candidateDirs := []string{
		fmt.Sprintf("base-%s", tag),
		fmt.Sprintf("base-%s", normalized),
	}

	var extractedDir string
	for _, d := range candidateDirs {
		p := filepath.Join(projectName, d)
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			extractedDir = p
			break
		}
	}

	if extractedDir == "" {
		if entries, err := os.ReadDir(projectName); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "base-") {
					extractedDir = filepath.Join(projectName, e.Name())
					break
				}
			}
		}
	}

	if extractedDir == "" {
		return fmt.Errorf("could not locate extracted template directory for tag %s", tag)
	}

	progress("Setting up project structure...")
	files, err := os.ReadDir(extractedDir)
	if err != nil {
		return fmt.Errorf("reading template directory: %w", err)
	}

	for _, f := range files {
		oldPath := filepath.Join(extractedDir, f.Name())
		newPath := filepath.Join(projectName, f.Name())
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("moving file %s: %w", f.Name(), err)
		}
	}

	os.RemoveAll(extractedDir)

	progress("Configuring environment...")
	if err := copyFile(filepath.Join(projectName, ".env.sample"), filepath.Join(projectName, ".env")); err != nil {
		// Non-fatal — .env.sample may not exist
		progress("Warning: could not copy .env.sample to .env")
	}

	progress("Project created successfully!")
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Chmod(dst, 0755)
}

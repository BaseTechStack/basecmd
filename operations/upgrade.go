package operations

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/base-go/cmd/version"
)

// UpgradeResult contains the result of an upgrade check.
type UpgradeResult struct {
	Release        *version.Release
	TargetVersion  string
	CurrentVersion string
	IsMajor        bool
	IsUpToDate     bool
}

// CheckUpgrade checks for available upgrades.
func CheckUpgrade(allowMajor, force bool) (*UpgradeResult, error) {
	release, targetVersion, err := getTargetVersion(allowMajor)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}

	info := version.GetBuildInfo()
	result := &UpgradeResult{
		Release:        release,
		TargetVersion:  targetVersion,
		CurrentVersion: info.Version,
		IsMajor:        IsMajorVersionUpgrade(info.Version, targetVersion),
		IsUpToDate:     info.Version == targetVersion && !force,
	}

	return result, nil
}

// PerformUpgrade downloads and installs a new version.
func PerformUpgrade(release *version.Release, targetVersion string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	osName := runtime.GOOS
	archName := runtime.GOARCH
	assetPrefix := fmt.Sprintf("base_%s_%s", osName, archName)
	var assetSuffix string
	if osName == "windows" {
		assetSuffix = ".zip"
	} else {
		assetSuffix = ".tar.gz"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, assetPrefix) && strings.HasSuffix(asset.Name, assetSuffix) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no compatible binary found for %s_%s", osName, archName)
	}

	progress(fmt.Sprintf("Downloading version %s...", targetVersion))

	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "base-new")

	if err := downloadAndInstall(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	progress("Verifying binary...")
	execCmd := exec.Command(tmpFile, "version")
	if err := execCmd.Run(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("verifying binary: %w", err)
	}

	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("making binary executable: %w", err)
	}

	var installDir, binDir, binaryName string
	if runtime.GOOS == "windows" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("getting home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".base")
		binDir = filepath.Join(homeDir, "bin")
		binaryName = "base.exe"
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("getting home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".base")
		binDir = "/usr/local/bin"
		binaryName = "base"
	}

	progress("Installing binary...")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("creating installation directory: %w", err)
	}
	if err := os.MkdirAll(binDir, 0755); err != nil && !os.IsExist(err) {
		os.Remove(tmpFile)
		return fmt.Errorf("creating bin directory: %w", err)
	}

	destPath := filepath.Join(installDir, binaryName)
	if err := os.Rename(tmpFile, destPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("moving binary: %w", err)
	}

	if runtime.GOOS == "windows" {
		if err := copyFile(destPath, filepath.Join(binDir, binaryName)); err != nil {
			return fmt.Errorf("copying binary to bin directory: %w", err)
		}
	} else {
		progress("Creating symlink (requires sudo)...")
		if err := runWithSudo("ln", "-sf", destPath, filepath.Join(binDir, binaryName)); err != nil {
			return fmt.Errorf("creating symlink - run manually: sudo ln -sf %s %s", destPath, filepath.Join(binDir, binaryName))
		}
	}

	progress(fmt.Sprintf("Successfully upgraded to version %s!", targetVersion))
	return nil
}

// IsMajorVersionUpgrade checks if the upgrade crosses a major version boundary.
func IsMajorVersionUpgrade(currentVersion, latestVersion string) bool {
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")
	if len(currentParts) == 0 || len(latestParts) == 0 {
		return false
	}
	return currentParts[0] != latestParts[0]
}

func getTargetVersion(allowMajor bool) (*version.Release, string, error) {
	release, err := version.CheckLatestVersion()
	if err != nil {
		return nil, "", err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	if !allowMajor {
		info := version.GetBuildInfo()
		currentVersion := strings.TrimPrefix(info.Version, "v")
		currentMajor := strings.Split(currentVersion, ".")[0]

		if strings.HasPrefix(latestVersion, currentMajor+".") {
			return release, latestVersion, nil
		}

		allReleases, err := version.GetAllReleases()
		if err != nil {
			return release, currentVersion, nil
		}

		latestInMajor := currentVersion
		for _, r := range allReleases {
			releaseVersion := strings.TrimPrefix(r.TagName, "v")
			if strings.HasPrefix(releaseVersion, currentMajor+".") {
				if version.CompareVersions(releaseVersion, latestInMajor) > 0 {
					latestInMajor = releaseVersion
				}
			}
		}
		return release, latestInMajor, nil
	}

	return release, latestVersion, nil
}

func extractTarGz(gzipStream io.Reader, targetPath string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			baseName := filepath.Base(header.Name)
			expectedName := "base"
			if runtime.GOOS == "windows" {
				expectedName = "base.exe"
			}

			if baseName == expectedName {
				outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
				if err != nil {
					return err
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, tarReader); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("binary not found in archive")
}

func extractZip(zipFile *os.File, targetPath string) error {
	stat, err := zipFile.Stat()
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return err
			}
			return os.Chmod(targetPath, 0755)
		}
	}
	return nil
}

func downloadAndInstall(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	tmpDir := os.TempDir()
	tmpArchive := filepath.Join(tmpDir, "base-archive")
	out, err := os.Create(tmpArchive)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		os.Remove(tmpArchive)
	}()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()

	archiveFile, err := os.Open(tmpArchive)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	if strings.HasSuffix(url, ".zip") {
		return extractZip(archiveFile, targetPath)
	} else if strings.HasSuffix(url, ".tar.gz") {
		return extractTarGz(archiveFile, targetPath)
	}

	return fmt.Errorf("unsupported archive format")
}

func runWithSudo(command string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{command}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

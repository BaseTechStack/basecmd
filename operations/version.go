package operations

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/version"
)

// VersionInfo holds formatted version information.
type VersionInfo struct {
	BuildInfo     version.BuildInfo
	UpdateAvail   bool
	LatestVersion string
	IsMajor       bool
	ReleaseURL    string
	ReleaseNotes  string
}

// GetVersionInfo returns complete version information with update check.
func GetVersionInfo() VersionInfo {
	info := version.GetBuildInfo()
	result := VersionInfo{BuildInfo: info}

	release, err := version.CheckLatestVersion()
	if err != nil {
		return result
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	result.LatestVersion = latestVersion
	result.ReleaseURL = release.HTMLURL
	result.ReleaseNotes = release.Body

	if version.HasUpdate(info.Version, latestVersion) {
		result.UpdateAvail = true
		result.IsMajor = IsMajorVersionUpgrade(info.Version, latestVersion)
	}

	return result
}

// FormatVersion returns a formatted string of version info.
func FormatVersion() string {
	info := version.GetBuildInfo()
	return fmt.Sprintf("Base CLI %s\nCommit: %s\nBuilt: %s\nGo version: %s",
		info.Version, info.CommitHash, info.BuildDate, info.GoVersion)
}

package operations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsMajorVersionUpgrade(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{"1.0.0", "1.1.0", false},
		{"1.0.0", "2.0.0", true},
		{"2.0.0", "2.1.0", false},
		{"1.5.3", "2.0.0", true},
		{"v1.0.0", "v2.0.0", true},
		{"v2.0.0", "v2.1.0", false},
		{"1.0.0", "1.0.0", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.current+"->"+tt.latest, func(t *testing.T) {
			got := IsMajorVersionUpgrade(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("IsMajorVersionUpgrade(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestFormatVersion(t *testing.T) {
	v := FormatVersion()
	if v == "" {
		t.Error("FormatVersion() returned empty string")
	}
	if len(v) < 10 {
		t.Errorf("FormatVersion() returned suspiciously short string: %q", v)
	}
}

func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()
	if info.BuildInfo.Version == "" {
		t.Error("GetVersionInfo().BuildInfo.Version is empty")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	content := []byte("hello world")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("copyFile content mismatch: got %q, want %q", got, content)
	}
}

func TestGetStringValue(t *testing.T) {
	m := map[string]interface{}{
		"name": "test",
		"num":  42.0,
	}
	if got := GetStringValue(m, "name"); got != "test" {
		t.Errorf("GetStringValue() = %q, want %q", got, "test")
	}
	if got := GetStringValue(m, "missing"); got != "" {
		t.Errorf("GetStringValue(missing) = %q, want empty", got)
	}
}

func TestGetBoolValue(t *testing.T) {
	m := map[string]interface{}{
		"active": true,
		"name":   "test",
	}
	if got := GetBoolValue(m, "active"); !got {
		t.Error("GetBoolValue(active) = false, want true")
	}
	if got := GetBoolValue(m, "name"); got {
		t.Error("GetBoolValue(name) = true, want false")
	}
	if got := GetBoolValue(m, "missing"); got {
		t.Error("GetBoolValue(missing) = true, want false")
	}
}

func TestGetIntValue(t *testing.T) {
	m := map[string]interface{}{
		"count":  42.0, // JSON numbers are float64
		"name":   "test",
		"actual": 7,
	}
	if got := GetIntValue(m, "count"); got != 42 {
		t.Errorf("GetIntValue(count) = %d, want 42", got)
	}
	if got := GetIntValue(m, "actual"); got != 7 {
		t.Errorf("GetIntValue(actual) = %d, want 7", got)
	}
	if got := GetIntValue(m, "missing"); got != 0 {
		t.Errorf("GetIntValue(missing) = %d, want 0", got)
	}
}

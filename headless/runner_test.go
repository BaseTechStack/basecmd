package headless

import (
	"testing"
)

func TestParseSchedulerFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantKey string
		wantURL string
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantKey: "",
			wantURL: "http://localhost:8100",
		},
		{
			name:    "api key with equals",
			args:    []string{"--api-key=test123"},
			wantKey: "test123",
			wantURL: "http://localhost:8100",
		},
		{
			name:    "url with equals",
			args:    []string{"--url=http://example.com"},
			wantKey: "",
			wantURL: "http://example.com",
		},
		{
			name:    "both flags separate",
			args:    []string{"--api-key", "key1", "--url", "http://custom:9000"},
			wantKey: "key1",
			wantURL: "http://custom:9000",
		},
		{
			name:    "flags mixed with positional args",
			args:    []string{"task-name", "--api-key=mykey"},
			wantKey: "mykey",
			wantURL: "http://localhost:8100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotURL := parseSchedulerFlags(tt.args)
			if gotKey != tt.wantKey {
				t.Errorf("apiKey = %q, want %q", gotKey, tt.wantKey)
			}
			if gotURL != tt.wantURL {
				t.Errorf("baseURL = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

func TestFilterFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "no flags",
			args: []string{"arg1", "arg2"},
			want: []string{"arg1", "arg2"},
		},
		{
			name: "flags with equals",
			args: []string{"--api-key=test", "arg1"},
			want: []string{"arg1"},
		},
		{
			name: "flags with space value",
			args: []string{"--api-key", "test", "arg1"},
			want: []string{"arg1"},
		},
		{
			name: "mixed flags and positional",
			args: []string{"task-name", "--api-key=key", "--url=http://localhost"},
			want: []string{"task-name"},
		},
		{
			name: "empty",
			args: []string{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterFlags(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d (got %v)", len(got), len(tt.want), got)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("filterFlags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRunUnknownCommand(t *testing.T) {
	err := Run([]string{"nonexistent-command"})
	if err == nil {
		t.Error("Run(unknown) should return error")
	}
}

func TestRunNoArgs(t *testing.T) {
	err := Run([]string{})
	if err == nil {
		t.Error("Run([]) should return error")
	}
}

func TestRunNewMissingArg(t *testing.T) {
	err := Run([]string{"new"})
	if err == nil {
		t.Error("Run(new) without project name should return error")
	}
}

func TestRunGenerateMissingArg(t *testing.T) {
	err := Run([]string{"generate"})
	if err == nil {
		t.Error("Run(generate) without module name should return error")
	}
}

func TestRunDestroyMissingArg(t *testing.T) {
	err := Run([]string{"destroy"})
	if err == nil {
		t.Error("Run(destroy) without module name should return error")
	}
}

func TestRunHelp(t *testing.T) {
	// Help should not return an error
	err := Run([]string{"help"})
	if err != nil {
		t.Errorf("Run(help) returned unexpected error: %v", err)
	}
}

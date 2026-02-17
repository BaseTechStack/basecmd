package operations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/base-go/cmd/utils"
)

// SchedulerGenerateTask generates a new scheduled task file in a module.
func SchedulerGenerateTask(moduleName, taskName string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current directory: %w", err)
	}

	mainGoPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		return fmt.Errorf("not in a Base Framework project directory")
	}

	var modulePath string
	if moduleName == "core" {
		modulePath = filepath.Join(cwd, "core", moduleName)
	} else {
		variations := []string{
			moduleName,
			strings.ToLower(moduleName),
			utils.PluralizeClient.Plural(strings.ToLower(moduleName)),
		}
		for _, variation := range variations {
			testPath := filepath.Join(cwd, "app", variation)
			if _, err := os.Stat(testPath); err == nil {
				modulePath = testPath
				break
			}
		}
		if modulePath == "" {
			modulePath = filepath.Join(cwd, "app", moduleName)
		}
	}

	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' does not exist at %s", moduleName, modulePath)
	}

	taskFileName := fmt.Sprintf("%s_task.go", utils.ToSnakeCase(taskName))
	taskFilePath := filepath.Join(modulePath, taskFileName)

	if _, err := os.Stat(taskFilePath); err == nil {
		return fmt.Errorf("task file already exists: %s", taskFilePath)
	}

	actualPackageName := filepath.Base(modulePath)
	taskContent := generateTaskContent(actualPackageName, taskName)

	if err := os.WriteFile(taskFilePath, []byte(taskContent), 0644); err != nil {
		return fmt.Errorf("could not create task file: %w", err)
	}

	// Run goimports if available
	_ = exec.Command("goimports", "-w", taskFilePath).Run()

	progress(fmt.Sprintf("Task '%s' generated in module '%s'", taskName, moduleName))
	return nil
}

func generateTaskContent(moduleName, taskName string) string {
	packageName := moduleName
	if moduleName == "core" {
		packageName = "core"
	}

	taskStructName := utils.ToPascalCase(taskName) + "Task"
	taskID := utils.ToKebabCase(taskName)

	return fmt.Sprintf(`package %s

import (
	"context"
	"time"

	"base/core/logger"
	"base/core/scheduler"
)

type %s struct {
	logger logger.Logger
}

func New%s(log logger.Logger) *%s {
	return &%s{
		logger: log,
	}
}

func (t *%s) RegisterTask(s *scheduler.Scheduler) error {
	task := &scheduler.Task{
		Name:        "%s",
		Description: "%s task for %s module",
		Schedule:    &scheduler.DailySchedule{Hour: 2, Minute: 0},
		Handler:     t.execute,
		Enabled:     true,
	}
	return s.RegisterTask(task)
}

func (t *%s) RegisterCronTask(cs *scheduler.CronScheduler) error {
	task := &scheduler.CronTask{
		Name:        "%s",
		Description: "%s task for %s module",
		CronExpr:    "0 0 2 * * *",
		Handler:     t.execute,
		Enabled:     true,
	}
	return cs.RegisterTask(task)
}

func (t *%s) execute(ctx context.Context) error {
	t.logger.Info("Starting %s task")

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// TODO: Implement your task logic here
	time.Sleep(1 * time.Second)

	t.logger.Info("%s task completed successfully")
	return nil
}

func (t *%s) GetTaskInfo() map[string]any {
	return map[string]any{
		"name":        "%s",
		"description": "%s task for %s module",
		"module":      "%s",
		"type":        "scheduled_task",
	}
}
`,
		packageName,
		taskStructName,
		taskStructName, taskStructName, taskStructName,
		taskStructName,
		taskID, taskName, moduleName,
		taskStructName,
		taskID, taskName, moduleName,
		taskStructName, taskName,
		taskName,
		taskStructName,
		taskID, taskName, moduleName, moduleName,
	)
}

// SchedulerAPIRequest makes an API request to the scheduler endpoint.
func SchedulerAPIRequest(method, endpoint, apiKey, baseURL string, body interface{}) (map[string]interface{}, error) {
	url := strings.TrimSuffix(baseURL, "/") + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("scheduler API endpoint not found (404)")
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// SchedulerListTasks lists tasks via API.
func SchedulerListTasks(apiKey, baseURL string) (map[string]interface{}, error) {
	return SchedulerAPIRequest("GET", "/api/scheduler/tasks", apiKey, baseURL, nil)
}

// SchedulerRunTask runs a specific task immediately.
func SchedulerRunTask(taskName, apiKey, baseURL string) (map[string]interface{}, error) {
	return SchedulerAPIRequest("POST", fmt.Sprintf("/api/scheduler/tasks/%s/run", taskName), apiKey, baseURL, nil)
}

// SchedulerEnableTask enables a task.
func SchedulerEnableTask(taskName, apiKey, baseURL string) (map[string]interface{}, error) {
	return SchedulerAPIRequest("PUT", fmt.Sprintf("/api/scheduler/tasks/%s/enable", taskName), apiKey, baseURL, nil)
}

// SchedulerDisableTask disables a task.
func SchedulerDisableTask(taskName, apiKey, baseURL string) (map[string]interface{}, error) {
	return SchedulerAPIRequest("PUT", fmt.Sprintf("/api/scheduler/tasks/%s/disable", taskName), apiKey, baseURL, nil)
}

// SchedulerGetStatus gets scheduler status via API.
func SchedulerGetStatus(apiKey, baseURL string) (map[string]interface{}, error) {
	return SchedulerAPIRequest("GET", "/api/scheduler/status", apiKey, baseURL, nil)
}

// GetStringValue extracts a string value from a map.
func GetStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetBoolValue extracts a bool value from a map.
func GetBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetIntValue extracts an int value from a map.
func GetIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

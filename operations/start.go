package operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StartServer starts the Base application development server.
// It returns the *exec.Cmd so callers can manage the process lifecycle.
func StartServer(withDocs bool, progress func(string)) (*exec.Cmd, error) {
	if progress == nil {
		progress = func(string) {}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}

	mainPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Base project structure not found (no main.go at %s)", mainPath)
	}

	goPath, err := findGo()
	if err != nil {
		return nil, err
	}

	progress("Ensuring dependencies are up to date...")
	tidyCmd := exec.Command(goPath, "mod", "tidy")
	tidyCmd.Dir = cwd
	_ = tidyCmd.Run()

	if withDocs {
		progress("Generating Swagger docs...")
		if err := generateSwaggerDocs(cwd, goPath); err != nil {
			progress(fmt.Sprintf("Warning: Swagger generation failed: %v", err))
		}
	}

	progress("Starting the Base application server...")
	mainCmd := exec.Command(goPath, "run", "main.go")
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	mainCmd.Dir = cwd

	env := os.Environ()
	if withDocs {
		env = append(env, "SWAGGER_ENABLED=true")
	}
	mainCmd.Env = env

	return mainCmd, nil
}

// RunServer is a convenience that starts and waits for the server.
func RunServer(withDocs bool, progress func(string)) error {
	cmd, err := StartServer(withDocs, progress)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func findGo() (string, error) {
	whichCmd := exec.Command("which", "go")
	goPathBytes, err := whichCmd.Output()
	if err != nil {
		return "", fmt.Errorf("Go executable not found: %w (ensure Go is in your PATH)", err)
	}
	return strings.TrimSpace(string(goPathBytes)), nil
}

func generateSwaggerDocs(cwd, goPath string) error {
	if _, err := exec.LookPath("swag"); err != nil {
		installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("installing swag: %w", err)
		}
	}

	swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./docs",
		"--parseDependency", "--parseInternal", "--parseVendor",
		"--parseDepth", "1", "--generatedTime", "false")
	swagCmd.Dir = cwd
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr
	return swagCmd.Run()
}

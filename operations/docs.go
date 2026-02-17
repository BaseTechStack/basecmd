package operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func GenerateDocs(outputDir string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}
	if outputDir == "" {
		outputDir = "docs"
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	mainPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		return fmt.Errorf("Base project structure not found (no main.go at %s)", mainPath)
	}

	goPath, err := findGo()
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("swag"); err != nil {
		progress("Installing swag...")
		installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("installing swag: %w", err)
		}
	}

	progress("Generating swagger documentation...")
	swagCmd := exec.Command("swag", "init",
		"--dir", "./",
		"--output", "./"+outputDir,
		"--parseDependency", "--parseInternal", "--parseVendor",
		"--parseDepth", "1", "--generatedTime", "false",
	)
	swagCmd.Dir = cwd
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr

	if err := swagCmd.Run(); err != nil {
		return fmt.Errorf("generating docs: %w", err)
	}

	progress(fmt.Sprintf("Swagger docs generated: %s/swagger.json, %s/swagger.yaml, %s/docs.go", outputDir, outputDir, outputDir))
	return nil
}

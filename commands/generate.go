package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/base-go/cmd/utils"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var generateCmd = &cobra.Command{
	Use:     "generate [name] [field:type...]",
	Aliases: []string{"g"},
	Short:   "Generate a new module",
	Long:    `Generate a new module with the specified name and fields.`,
	Args:    cobra.MinimumNArgs(1),
	Run:     generateModule,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

// generateModule generates a new module with the specified name and fields.
func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Create naming convention from the input name
	naming := utils.NewNamingConvention(singularName)

	// Create directories (plural names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", naming.DirName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	// Generate field structs
	fieldStructs := utils.NewTemplateData(naming.Model, fields)

	// Generate model
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		naming.ModelSnake+".go",
		"model.tmpl",
		naming,
		fieldStructs.Fields,
	)

	// Generate service
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"service.go",
		"service.tmpl",
		naming,
		fieldStructs.Fields,
	)

	// Generate controller
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"controller.go",
		"controller.tmpl",
		naming,
		fieldStructs.Fields,
	)

	// Generate module
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"module.go",
		"module.tmpl",
		naming,
		fieldStructs.Fields,
	)

	// Generate validator
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"validator.go",
		"validator.tmpl",
		naming,
		fieldStructs.Fields,
	)

	// Generate tests - disabled for now, will be added in future
	// if err := utils.GenerateTests(naming, fieldStructs); err != nil {
	// 	fmt.Printf("Error generating tests: %v\n", err)
	// 	return
	// }

	// Check if goimports is installed
	if _, err := exec.LookPath("goimports"); err != nil {
		fmt.Println("goimports not found, installing...")
		if err := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").Run(); err != nil {
			fmt.Printf("Error installing goimports: %v\n", err)
			fmt.Println("Please install goimports manually: go install golang.org/x/tools/cmd/goimports@latest")
			return
		}
		fmt.Println("goimports installed successfully")
	}

	// Run goimports on generated files
	generatedPath := filepath.Join("app", naming.DirName)

	fmt.Println("Running goimports on generated files...")
	// Run goimports on the generated directory
	if err := exec.Command("find", generatedPath, "-name", "*.go", "-exec", "goimports", "-w", "{}", ";").Run(); err != nil {
		fmt.Printf("Error running goimports on %s: %v\n", generatedPath, err)
	}

	// Run goimports on the model file
	modelPath := filepath.Join("app", "models", naming.ModelSnake+".go")
	if err := exec.Command("goimports", "-w", modelPath).Run(); err != nil {
		fmt.Printf("Error running goimports on %s: %v\n", modelPath, err)
	}

	// Format all generated files with gofmt
	fmt.Println("Formatting generated files...")
	if err := exec.Command("gofmt", "-w", generatedPath).Run(); err != nil {
		fmt.Printf("Warning: Failed to format generated files in %s: %v\n", generatedPath, err)
	}
	if err := exec.Command("gofmt", "-w", modelPath).Run(); err != nil {
		fmt.Printf("Warning: Failed to format model file %s: %v\n", modelPath, err)
	}

	// Add module to app/init.go
	if err := addModuleToAppInit(naming.DirName); err != nil {
		fmt.Printf("Warning: Could not add module to app/init.go: %v\n", err)
		fmt.Printf("Please manually add: _ \"base/app/%s\" to app/init.go\n", naming.DirName)
	} else {
		fmt.Printf("✅ Added module to app/init.go\n")

		// Format init.go after modification
		initGoPath := filepath.Join("app", "init.go")
		if err := exec.Command("gofmt", "-w", initGoPath).Run(); err != nil {
			fmt.Printf("Warning: Failed to format app/init.go: %v\n", err)
		} else {
			fmt.Printf("✅ Formatted app/init.go\n")
		}
	}

	// Run go mod tidy to ensure dependencies are up to date
	fmt.Println("Running go mod tidy...")
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		fmt.Printf("Warning: Failed to run go mod tidy: %v\n", err)
	} else {
		fmt.Printf("✅ Dependencies updated\n")
	}

	fmt.Printf("Successfully generated %s module\n", naming.Model)
}

// addModuleToAppInit adds the module to app/init.go
func addModuleToAppInit(moduleName string) error {
	initGoPath := filepath.Join("app", "init.go")

	// Check if app/init.go exists
	if _, err := os.Stat(initGoPath); os.IsNotExist(err) {
		// Create app/init.go if it doesn't exist
		if err := os.MkdirAll("app", os.ModePerm); err != nil {
			return fmt.Errorf("failed to create app directory: %w", err)
		}

		content := fmt.Sprintf(`package app

import (
	"base/app/%s"
	"base/core/module"
)

// AppModules implements module.AppModuleProvider interface
type AppModules struct{}

// GetAppModules returns the list of app modules to initialize
// This is the only function that needs to be updated when adding new app modules
func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	// App modules - custom system functionality
	modules["%s"] = %s.Init(deps)

	return modules
}

// NewAppModules creates a new AppModules provider
func NewAppModules() *AppModules {
	return &AppModules{}
}
`, moduleName, moduleName, moduleName)

		if err := os.WriteFile(initGoPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create app/init.go: %w", err)
		}
		return nil
	}

	// Read existing app/init.go content
	content, err := os.ReadFile(initGoPath)
	if err != nil {
		return fmt.Errorf("failed to read app/init.go: %w", err)
	}

	contentStr := string(content)

	// Check if module already exists
	moduleInit := fmt.Sprintf("modules[\"%s\"] = %s.Init(deps)", moduleName, moduleName)
	if strings.Contains(contentStr, moduleInit) {
		return nil // Already added
	}

	// Add import if not exists using the proper AddImport function
	importLine := fmt.Sprintf("\"base/app/%s\"", moduleName)
	contentBytes, importAdded := utils.AddImport([]byte(contentStr), importLine)
	if importAdded {
		contentStr = string(contentBytes)
	}

	// Add module initialization
	// Find the return modules line
	returnIndex := strings.Index(contentStr, "return modules")
	if returnIndex == -1 {
		return fmt.Errorf("could not find 'return modules' in app/init.go")
	}

	// Insert the module initialization before return
	insertPoint := returnIndex - 1
	caser := cases.Title(language.English)
	moduleInitLine := fmt.Sprintf("\n\t// %s module\n\t%s\n", caser.String(moduleName), moduleInit)
	contentStr = contentStr[:insertPoint] + moduleInitLine + contentStr[insertPoint:]

	// Write back to file
	if err := os.WriteFile(initGoPath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write app/init.go: %w", err)
	}

	return nil
}

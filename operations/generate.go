package operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/base-go/cmd/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateModule(name string, fields []string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	naming := utils.NewNamingConvention(name)

	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", naming.DirName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	progress("Generating module files...")
	fieldStructs := utils.NewTemplateData(naming.Model, fields)

	templates := []struct {
		dir, filename, tmpl string
	}{
		{filepath.Join("app", "models"), naming.ModelSnake + ".go", "model.tmpl"},
		{filepath.Join("app", naming.DirName), "service.go", "service.tmpl"},
		{filepath.Join("app", naming.DirName), "controller.go", "controller.tmpl"},
		{filepath.Join("app", naming.DirName), "module.go", "module.tmpl"},
		{filepath.Join("app", naming.DirName), "validator.go", "validator.tmpl"},
	}

	for _, t := range templates {
		utils.GenerateFileFromTemplate(t.dir, t.filename, t.tmpl, naming, fieldStructs.Fields)
	}

	progress("Formatting code...")
	ensureGoimports()

	generatedPath := filepath.Join("app", naming.DirName)
	_ = exec.Command("find", generatedPath, "-name", "*.go", "-exec", "goimports", "-w", "{}", ";").Run()

	modelPath := filepath.Join("app", "models", naming.ModelSnake+".go")
	_ = exec.Command("goimports", "-w", modelPath).Run()
	_ = exec.Command("gofmt", "-w", generatedPath).Run()
	_ = exec.Command("gofmt", "-w", modelPath).Run()

	progress("Registering module...")
	if err := addModuleToAppInit(naming.DirName); err != nil {
		progress(fmt.Sprintf("Warning: Could not add module to app/init.go: %v", err))
	} else {
		initGoPath := filepath.Join("app", "init.go")
		_ = exec.Command("gofmt", "-w", initGoPath).Run()
	}

	progress("Updating dependencies...")
	_ = exec.Command("go", "mod", "tidy").Run()

	progress(fmt.Sprintf("Module '%s' generated successfully!", naming.Model))
	return nil
}

func ensureGoimports() {
	if _, err := exec.LookPath("goimports"); err != nil {
		_ = exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").Run()
	}
}

func addModuleToAppInit(moduleName string) error {
	initGoPath := filepath.Join("app", "init.go")

	if _, err := os.Stat(initGoPath); os.IsNotExist(err) {
		if err := os.MkdirAll("app", os.ModePerm); err != nil {
			return fmt.Errorf("failed to create app directory: %w", err)
		}
		content := fmt.Sprintf(`package app

import (
	"base/app/%s"
	"base/core/module"
)

type AppModules struct{}

func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)
	modules["%s"] = %s.Init(deps)
	return modules
}

func NewAppModules() *AppModules {
	return &AppModules{}
}
`, moduleName, moduleName, moduleName)
		return os.WriteFile(initGoPath, []byte(content), 0644)
	}

	content, err := os.ReadFile(initGoPath)
	if err != nil {
		return fmt.Errorf("failed to read app/init.go: %w", err)
	}

	contentStr := string(content)

	moduleInit := fmt.Sprintf("modules[\"%s\"] = %s.Init(deps)", moduleName, moduleName)
	if strings.Contains(contentStr, moduleInit) {
		return nil
	}

	importLine := fmt.Sprintf("\"base/app/%s\"", moduleName)
	contentBytes, importAdded := utils.AddImport([]byte(contentStr), importLine)
	if importAdded {
		contentStr = string(contentBytes)
	}

	returnIndex := strings.Index(contentStr, "return modules")
	if returnIndex == -1 {
		return fmt.Errorf("could not find 'return modules' in app/init.go")
	}

	insertPoint := returnIndex - 1
	caser := cases.Title(language.English)
	moduleInitLine := fmt.Sprintf("\n\t// %s module\n\t%s\n", caser.String(moduleName), moduleInit)
	contentStr = contentStr[:insertPoint] + moduleInitLine + contentStr[insertPoint:]

	return os.WriteFile(initGoPath, []byte(contentStr), 0644)
}

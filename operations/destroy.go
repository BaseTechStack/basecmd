package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/base-go/cmd/utils"
)

// DestroyModules destroys the given module names. Returns an error if any fail.
func DestroyModules(moduleNames []string, progress func(string)) error {
	if progress == nil {
		progress = func(string) {}
	}

	allOk := true
	for i, name := range moduleNames {
		progress(fmt.Sprintf("[%d/%d] Destroying module '%s'...", i+1, len(moduleNames), name))
		if !destroySingleModule(name, progress) {
			allOk = false
		}
	}

	if !allOk {
		return fmt.Errorf("some modules could not be fully destroyed")
	}
	return nil
}

func destroySingleModule(singularName string, progress func(string)) bool {
	pluralName := utils.ToSnakeCase(utils.ToPlural(singularName))
	singularDirName := utils.ToSnakeCase(singularName)

	var moduleDir string
	var moduleExists bool
	success := true

	pluralDir := filepath.Join("app", pluralName)
	singularDir := filepath.Join("app", singularDirName)

	if _, err := os.Stat(pluralDir); err == nil {
		moduleDir = pluralDir
		moduleExists = true
	} else if _, err := os.Stat(singularDir); err == nil {
		moduleDir = singularDir
		moduleExists = true
	}

	if moduleExists {
		if err := os.RemoveAll(moduleDir); err != nil {
			progress(fmt.Sprintf("  Error removing directory %s: %v", moduleDir, err))
			success = false
		} else {
			progress(fmt.Sprintf("  Removed module directory: %s", moduleDir))
		}
	}

	if moduleExists {
		modelFiles := []string{
			filepath.Join("app", "models", utils.ToSnakeCase(singularName)+".go"),
			filepath.Join("app", "model", utils.ToSnakeCase(singularName)+".go"),
		}

		for _, modelFile := range modelFiles {
			if err := os.Remove(modelFile); err == nil {
				progress(fmt.Sprintf("  Removed model file: %s", modelFile))
				break
			}
		}
	}

	if moduleExists {
		testDir := filepath.Join("test", "app_test", pluralName+"_test")
		if _, err := os.Stat(testDir); err == nil {
			if err := os.RemoveAll(testDir); err != nil {
				progress(fmt.Sprintf("  Warning: Could not remove test directory %s: %v", testDir, err))
			} else {
				progress(fmt.Sprintf("  Removed test directory: %s", testDir))
			}
		}
	}

	if err := removeModuleFromAppInit(pluralName); err != nil {
		progress(fmt.Sprintf("  Warning: Could not remove from app/init.go: %v", err))
	} else {
		progress(fmt.Sprintf("  Removed '%s' from app/init.go", pluralName))
	}

	return success
}

func removeModuleFromAppInit(moduleName string) error {
	initGoPath := filepath.Join("app", "init.go")

	if _, err := os.Stat(initGoPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(initGoPath)
	if err != nil {
		return fmt.Errorf("failed to read app/init.go: %w", err)
	}

	importStr := fmt.Sprintf("\"base/app/%s\"", moduleName)
	content = utils.RemoveImport(content, importStr)
	content = utils.RemoveModuleInitializer(content, moduleName)

	return os.WriteFile(initGoPath, content, 0644)
}

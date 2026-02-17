package headless

import (
	"fmt"
	"os"
	"strings"

	"github.com/base-go/cmd/operations"
	"github.com/base-go/cmd/version"
)

// Run handles CLI arguments for CI/scripting (headless) mode.
func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	rest := args[1:]

	// Check for updates before most commands (like Cobra's PersistentPreRun)
	if command != "version" && command != "upgrade" {
		checkForUpdates()
	}

	switch command {
	case "new":
		return runNew(rest)
	case "generate", "g":
		return runGenerate(rest)
	case "start", "s":
		return runStart(rest)
	case "docs":
		return runDocs(rest)
	case "d", "destroy":
		return runDestroy(rest)
	case "scheduler", "sc":
		return runScheduler(rest)
	case "update":
		return runUpdate()
	case "upgrade":
		return runUpgrade(rest)
	case "version":
		return runVersion()
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func checkForUpdates() {
	release, err := version.CheckLatestVersion()
	if err != nil {
		return
	}
	info := version.GetBuildInfo()
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if version.HasUpdate(info.Version, latestVersion) {
		fmt.Print(version.FormatUpdateMessage(
			info.Version,
			latestVersion,
			release.HTMLURL,
			release.Body,
		))
	}
}

func printUsage() {
	fmt.Println("Base CLI - A modern Go web framework")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  base                         Open interactive TUI")
	fmt.Println("  base <command> [args]         Run command directly")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  new <name>                   Create a new Base project")
	fmt.Println("  generate <name> [fields...]  Generate a new module (alias: g)")
	fmt.Println("  start [--docs]               Start the development server (alias: s)")
	fmt.Println("  docs [--output <dir>]        Generate Swagger documentation")
	fmt.Println("  destroy <name> [names...]    Destroy existing modules (alias: d)")
	fmt.Println("  scheduler <subcommand>       Manage scheduled tasks (alias: sc)")
	fmt.Println("  update                       Update Base Core to latest version")
	fmt.Println("  upgrade [--major] [--force]  Upgrade Base CLI")
	fmt.Println("  version                      Print version information")
	fmt.Println("  help                         Show this help message")
}

func progress(s string) {
	fmt.Println(s)
}

func runNew(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: base new <project_name>")
	}
	return operations.NewProject(args[0], progress)
}

func runGenerate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: base generate <name> [field:type...]")
	}
	return operations.GenerateModule(args[0], args[1:], progress)
}

func runStart(args []string) error {
	withDocs := false
	for _, a := range args {
		if a == "--docs" || a == "-d" {
			withDocs = true
		}
	}
	return operations.RunServer(withDocs, progress)
}

func runDocs(args []string) error {
	outputDir := "docs"
	for i, a := range args {
		if (a == "--output" || a == "-o") && i+1 < len(args) {
			outputDir = args[i+1]
		}
	}
	return operations.GenerateDocs(outputDir, progress)
}

func runDestroy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: base destroy <name> [names...]")
	}

	// Confirmation prompt
	fmt.Printf("Modules to destroy: %s\n", strings.Join(args, ", "))
	fmt.Printf("Are you sure you want to destroy %d module(s)? [Y/n] ", len(args))
	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	return operations.DestroyModules(args, progress)
}

func runScheduler(args []string) error {
	if len(args) < 1 {
		fmt.Println("Scheduler subcommands:")
		fmt.Println("  generate <module> <task>  Generate a new scheduled task")
		fmt.Println("  list                      List all registered tasks")
		fmt.Println("  run <task>                Run a specific task immediately")
		fmt.Println("  enable <task>             Enable a specific task")
		fmt.Println("  disable <task>            Disable a specific task")
		fmt.Println("  status                    Get scheduler status")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --api-key <key>           API key for authentication")
		fmt.Println("  --url <url>               Base URL (default: http://localhost:8100)")
		return nil
	}

	sub := args[0]
	rest := args[1:]

	apiKey, baseURL := parseSchedulerFlags(rest)

	switch sub {
	case "generate", "g":
		positional := filterFlags(rest)
		if len(positional) < 2 {
			return fmt.Errorf("usage: base scheduler generate <module> <task-name>")
		}
		return operations.SchedulerGenerateTask(positional[0], positional[1], progress)

	case "list", "ls":
		if apiKey != "" {
			tasks, err := operations.SchedulerListTasks(apiKey, baseURL)
			if err != nil {
				return err
			}
			printTaskList(tasks)
		} else {
			fmt.Println("Provide --api-key to list tasks from running server.")
		}
		return nil

	case "run":
		positional := filterFlags(rest)
		if len(positional) < 1 {
			return fmt.Errorf("usage: base scheduler run <task-name>")
		}
		if apiKey != "" {
			_, err := operations.SchedulerRunTask(positional[0], apiKey, baseURL)
			if err != nil {
				return err
			}
			fmt.Printf("Task '%s' executed successfully\n", positional[0])
		} else {
			fmt.Printf("Provide --api-key to run task '%s'\n", positional[0])
		}
		return nil

	case "enable":
		positional := filterFlags(rest)
		if len(positional) < 1 {
			return fmt.Errorf("usage: base scheduler enable <task-name>")
		}
		if apiKey != "" {
			_, err := operations.SchedulerEnableTask(positional[0], apiKey, baseURL)
			if err != nil {
				return err
			}
			fmt.Printf("Task '%s' enabled\n", positional[0])
		} else {
			fmt.Printf("Provide --api-key to enable task '%s'\n", positional[0])
		}
		return nil

	case "disable":
		positional := filterFlags(rest)
		if len(positional) < 1 {
			return fmt.Errorf("usage: base scheduler disable <task-name>")
		}
		if apiKey != "" {
			_, err := operations.SchedulerDisableTask(positional[0], apiKey, baseURL)
			if err != nil {
				return err
			}
			fmt.Printf("Task '%s' disabled\n", positional[0])
		} else {
			fmt.Printf("Provide --api-key to disable task '%s'\n", positional[0])
		}
		return nil

	case "status":
		if apiKey != "" {
			status, err := operations.SchedulerGetStatus(apiKey, baseURL)
			if err != nil {
				return err
			}
			printSchedulerStatus(status)
		} else {
			fmt.Println("Provide --api-key to get scheduler status.")
		}
		return nil

	default:
		return fmt.Errorf("unknown scheduler subcommand: %s", sub)
	}
}

func parseSchedulerFlags(args []string) (apiKey, baseURL string) {
	baseURL = "http://localhost:8100"
	for i, a := range args {
		if strings.HasPrefix(a, "--api-key=") {
			apiKey = strings.TrimPrefix(a, "--api-key=")
		} else if a == "--api-key" && i+1 < len(args) {
			apiKey = args[i+1]
		}
		if strings.HasPrefix(a, "--url=") {
			baseURL = strings.TrimPrefix(a, "--url=")
		} else if a == "--url" && i+1 < len(args) {
			baseURL = args[i+1]
		}
	}
	return
}

func filterFlags(args []string) []string {
	var positional []string
	skip := false
	for _, a := range args {
		if skip {
			skip = false
			continue
		}
		if strings.HasPrefix(a, "--") {
			if !strings.Contains(a, "=") {
				skip = true // next arg is the flag value
			}
			continue
		}
		positional = append(positional, a)
	}
	return positional
}

func printTaskList(resp map[string]interface{}) {
	if data, ok := resp["data"]; ok {
		if tasks, ok := data.([]interface{}); ok {
			fmt.Printf("\nFound %d tasks:\n\n", len(tasks))
			for _, taskInterface := range tasks {
				if task, ok := taskInterface.(map[string]interface{}); ok {
					name := operations.GetStringValue(task, "name")
					desc := operations.GetStringValue(task, "description")
					enabled := operations.GetBoolValue(task, "enabled")
					schedule := operations.GetStringValue(task, "schedule")
					runCount := operations.GetIntValue(task, "run_count")
					errorCount := operations.GetIntValue(task, "error_count")

					status := "Disabled"
					if enabled {
						status = "Enabled"
					}

					fmt.Printf("  %s [%s]\n", name, status)
					fmt.Printf("    %s\n", desc)
					fmt.Printf("    Schedule: %s | Runs: %d | Errors: %d\n\n", schedule, runCount, errorCount)
				}
			}
		}
	}
}

func printSchedulerStatus(resp map[string]interface{}) {
	if data, ok := resp["data"]; ok {
		if status, ok := data.(map[string]interface{}); ok {
			running := operations.GetBoolValue(status, "running")
			total := operations.GetIntValue(status, "total_tasks")
			enabled := operations.GetIntValue(status, "enabled_tasks")
			disabled := operations.GetIntValue(status, "disabled_tasks")

			runStr := "Stopped"
			if running {
				runStr = "Running"
			}

			fmt.Printf("\nScheduler: %s\n", runStr)
			fmt.Printf("Total: %d | Enabled: %d | Disabled: %d\n", total, enabled, disabled)
		}
	}
}

func runUpdate() error {
	fmt.Println("Updating Base Core...")
	if err := operations.UpdateCore(progress); err != nil {
		return err
	}
	fmt.Println("Base Core updated successfully.")
	return nil
}

func runUpgrade(args []string) error {
	allowMajor := false
	force := false
	for _, a := range args {
		if a == "--major" {
			allowMajor = true
		}
		if a == "--force" {
			force = true
		}
	}

	fmt.Println("Checking for updates...")
	result, err := operations.CheckUpgrade(allowMajor, force)
	if err != nil {
		return err
	}

	if result.IsUpToDate {
		fmt.Printf("You're already using the latest version (%s)\n", result.CurrentVersion)
		return nil
	}

	if result.IsMajor {
		fmt.Printf("\nMAJOR VERSION UPGRADE: %s -> %s\n", result.CurrentVersion, result.TargetVersion)
		fmt.Println("This may contain breaking changes!")
		fmt.Print("Do you want to proceed? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	return operations.PerformUpgrade(result.Release, result.TargetVersion, progress)
}

func runVersion() error {
	info := operations.GetVersionInfo()
	fmt.Println(info.BuildInfo.String())

	if info.UpdateAvail {
		if info.IsMajor {
			fmt.Printf("\nMAJOR VERSION AVAILABLE: %s -> %s\n", info.BuildInfo.Version, info.LatestVersion)
		} else {
			fmt.Print(version.FormatUpdateMessage(
				info.BuildInfo.Version,
				info.LatestVersion,
				info.ReleaseURL,
				info.ReleaseNotes,
			))
		}
	} else {
		fmt.Printf("\nYou're up to date! Using the latest version %s\n", info.BuildInfo.Version)
	}
	return nil
}

// ScanModules scans the app/ directory and returns module names.
func ScanModules() []string {
	var modules []string
	entries, err := os.ReadDir("app")
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != "models" && e.Name() != "model" {
			modules = append(modules, e.Name())
		}
	}
	return modules
}

// Package main provides a CLI tool for setting up and managing an OpenAI Responses API alternative
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// CLAUDE-note-data-structs: Core data structures for environment configuration and CLI setup (~30 lines)
// EnvVar represents an environment variable with metadata for setup prompts
type EnvVar struct {
	Name        string // Environment variable name
	Description string // Human-readable description
	Default     string // Default value
	Required    bool   // Whether this env var is required during setup
}

// Config represents the CLI configuration stored in JSON format
type Config struct {
	Version        string            `json:"version"`              // CLI version
	CreatedAt      string            `json:"createdAt,omitempty"`  // Timestamp when config was created
	UpdatedAt      string            `json:"updatedAt,omitempty"`  // Timestamp when config was last updated
	CreatedTime    string            `json:"created_at,omitempty"` // Alternate timestamp field for creation
	UpdatedTime    string            `json:"updated_at,omitempty"` // Alternate timestamp field for update
	Host           string            `json:"host"`                 // API host address
	Port           string            `json:"port"`                 // API port
	DockerTag      string            `json:"docker_tag"`           // Docker image tag
	BaseComposeURI string            `json:"base_compose_uri"`     // URL or path to base compose file
	EnvFile        string            `json:"env_file"`             // Path to env file
	APIVersion     string            `json:"api_version"`          // API version in semver format
	Environment    map[string]string `json:"environment"`          // Environment variable values
}

const (
	configFileName = "open-responses.json" // Name of the configuration file
	version        = "0.1.4"               // Current CLI version
)

// CLAUDE-note-root-cmd: Command structure definitions - Root command and CLI entrypoint (~23 lines)
// Root command for the CLI application
var rootCmd = &cobra.Command{
	Use:   "open-responses",
	Short: "A CLI for setting up an OpenAI Responses API alternative",
	Long:  `This CLI helps you set up and manage a Docker Compose service that mimics OpenAI's Responses API.`,

	// PersistentPreRunE runs before any command execution
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config check for setup and init commands
		if cmd.Use == "setup" || cmd.Use == "init" {
			return nil
		}

		// Check if this is the first run by looking for config file
		configExists, _ := checkConfigExists()
		if !configExists {
			return fmt.Errorf("no configuration found. Please run 'open-responses setup' first")
		}
		return nil
	},
}

// CLAUDE-note-compose-group: Docker Compose command group for organizing Docker Compose related commands
// Compose command group to contain all Docker Compose related commands
var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Docker Compose related commands for the Responses API service",
	Long:  `This group of commands contains all operations related to Docker Compose for managing the Responses API service.`,
}

// CLAUDE-note-root-config-cmd: Root configuration command for viewing and setting configuration values
// Root-level configuration command for viewing and managing the application configuration
var rootConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify configuration settings",
	Long: `The config command allows you to view or modify the configuration settings for the Responses API service.

Example:
  # View the current configuration:
  open-responses config

  # View a specific configuration value:
  open-responses config --show env_file`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

// CLAUDE-note-user-friendly-cmds: User-friendly root commands for common operations

// startCmd provides a simplified interface for starting the services
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Responses API service (alias for 'compose up -d')",
	Long: `Start the Responses API service in detached mode.

This is a user-friendly alias for 'open-responses compose up -d' that starts
all services in the background with sensible defaults. The command will:

1. Start all services defined in responses-compose.yaml
2. Run in detached mode (services run in the background)
3. Show the status of running services after startup
4. Display access URLs for the API and admin UI

For more advanced options, use 'open-responses compose up' with additional flags.

Example:
  # Start all services in the background:
  open-responses start`,
	Run: func(cmd *cobra.Command, args []string) {
		startService()
	},
}

// stopCmd provides a simplified interface for stopping the services
var rootStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Responses API service (alias for 'compose down')",
	Long: `Stop the Responses API service and clean up resources.

This is a user-friendly alias for 'open-responses compose down' that stops
all services and performs basic cleanup. The command will:

1. Stop all running containers
2. Remove containers and networks created by 'start'
3. Show confirmation when services are successfully stopped

For more advanced options, use 'open-responses compose down' with additional flags.

Example:
  # Stop all services and clean up:
  open-responses stop`,
	Run: func(cmd *cobra.Command, args []string) {
		stopRootService()
	},
}

// statusCmd provides detailed status information about the services
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all services",
	Long: `Display detailed information about the Responses API service status.

This command provides a user-friendly overview of all services including:

1. Running state of each container (running, stopped, or exited)
2. Health status (healthy, unhealthy, or starting)
3. Uptime information
4. Access URLs for the API and admin UI
5. Resource usage summary

Example:
  # Show the current service status:
  open-responses status`,
	Run: func(cmd *cobra.Command, args []string) {
		showServiceStatus()
	},
}

// logsCmd provides a simplified interface for viewing service logs
var rootLogsCmd = &cobra.Command{
	Use:   "logs [SERVICE]",
	Short: "View logs from services (simplified 'compose logs')",
	Long: `View logs from the Responses API services with sensible defaults.

This is a user-friendly version of 'open-responses compose logs' that:

1. Shows logs in follow mode by default (-f)
2. Displays colorized output for better readability
3. Can target a specific service or show all services
4. Uses sensible defaults for log formatting

Examples:
  # Follow logs from all services:
  open-responses logs

  # Follow logs from the API service only:
  open-responses logs api`,
	Run: func(cmd *cobra.Command, args []string) {
		showServiceLogs(args)
	},
}

// initCmd provides a streamlined setup for new projects
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Responses API project",
	Long: `Initialize a new Responses API project with guided setup.

This command is optimized for new project creation and will:

1. Create a new project structure in the current directory
2. Guide you through configuration with interactive prompts
3. Generate all necessary configuration files
4. Set up Docker Compose with best practice defaults
5. Configure API keys and environment variables
6. Create helpful documentation files

Example:
  # Initialize a new project in the current directory:
  open-responses init`,
	Run: func(cmd *cobra.Command, args []string) {
		initProject()
	},
}

// updateCmd provides functionality to update components
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Responses API components",
	Long: `Update the Responses API components to the latest version.

This command handles various update operations including:

1. Pulling the latest Docker images
2. Updating the Docker Compose template
3. Checking for CLI updates
4. Backing up your configuration before updates
5. Applying migrations when needed

Examples:
  # Update all components:
  open-responses update

  # Update only Docker images:
  open-responses update --images-only`,
	Run: func(cmd *cobra.Command, args []string) {
		updateComponents()
	},
}

// keyCmd provides API key management
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage API keys for the Responses API",
	Long: `Manage API keys for the Responses API service.

This command provides functionality to:

1. View existing API keys (partially masked for security)
2. Generate new API keys with proper format
3. Update API keys in your configuration
4. Rotate API keys for improved security

Examples:
  # View your current API keys (masked):
  open-responses key list

  # Generate a new Responses API key:
  open-responses key generate

  # Update the OpenAI API key:
  open-responses key set openai`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a key action: list, generate, set")
			os.Exit(1)
		}
		manageKeys(args)
	},
}

// CLAUDE-note-setup-cmd: Setup command - Primary configuration initialization workflow (~10 lines)
// Setup command for initializing the service configuration
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up the Responses API service configuration",
	Run: func(cmd *cobra.Command, args []string) {
		setupConfig()
	},
}

// CLAUDE-note-lifecycle-cmds: Service lifecycle management commands - Up, Stop, Down for controlling services (~84 lines)
// Up command for launching the Responses API service
// This is a direct proxy to `docker compose up` with argument passthrough
// Equivalent to Docker Compose's "up" command that creates and starts containers
var upCmd = &cobra.Command{
	Use:   "up [flags]",
	Short: "Start the Responses API service (docker compose up)",
	Long: `Start the Responses API service using Docker Compose.

This command is a direct proxy to 'docker compose up' and accepts all the same flags.
Common flags include:
  -d, --detach       Run containers in the background
  --build            Build images before starting containers
  --no-build         Don't build images, even if they're missing
  --force-recreate   Recreate containers even if their configuration hasn't changed

Examples:
  # Start in the foreground with output streaming to the console:
  open-responses compose up

  # Start in detached mode (background):
  open-responses compose up -d

  # Force rebuild of all services:
  open-responses compose up --build --force-recreate`,
	Run: func(cmd *cobra.Command, args []string) {
		upService(args)
	},
}

// Stop command for shutting down the service
// This is a legacy alias for "down" provided for backward compatibility
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Responses API service",
	Long: `Stop the Responses API service containers.
This is a legacy command provided for backward compatibility.
For complete cleanup, prefer the 'compose down' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		stopService()
	},
}

// Down command for shutting down the service (Docker Compose down)
// This is a direct proxy to `docker compose down`
// Equivalent to Docker Compose's "down" command that stops and removes containers,
// networks, and optionally volumes
var downCmd = &cobra.Command{
	Use:   "down [flags]",
	Short: "Stop the Responses API service and remove containers (docker compose down)",
	Long: `Stop the Responses API service and remove containers, networks, and volumes
defined in responses-compose.yaml.

This command is a direct proxy to 'docker compose down'. It stops and removes
containers, networks, and optionally volumes created by 'up'.

Common flags include:
  -v, --volumes           Remove named volumes declared in the volumes section
  --rmi string            Remove images. Type must be one of: 'all': all images, 'local': only locally built images
  --remove-orphans        Remove containers for services not defined in the Compose file

Examples:
  # Stop and remove containers and networks:
  open-responses compose down

  # Stop and remove containers, networks, and volumes:
  open-responses compose down -v

  # Remove all images used by services:
  open-responses compose down --rmi all`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get configuration path for informational output
		_, configPath := checkConfigExists()

		// Inform the user which configuration we're using
		fmt.Printf("Stopping services with configuration from %s\n", configPath)

		// Execute Docker Compose "down" command with all args
		executeDockerComposeWithArgs("down", args)

		fmt.Println("Service stopped successfully!")
	},
}

// CLAUDE-note-monitor-cmds: Service monitoring commands - Logs, PS, Top, Stats for observing service state (~138 lines)
// Logs command for viewing container logs
// This is a direct proxy to `docker compose logs` with argument passthrough
var logsCmd = &cobra.Command{
	Use:   "logs [flags] [SERVICE...]",
	Short: "View output from containers (docker compose logs)",
	Long: `View output from containers.

This command is a direct proxy to 'docker compose logs' and accepts all the same flags.
Common flags include:
  -f, --follow       Follow log output
  --tail string      Number of lines to show from the end of the logs (default "all")
  --since string     Show logs since timestamp (e.g. "2013-01-02T13:23:37Z")
  --until string     Show logs before timestamp

Examples:
  # Show logs from all services:
  open-responses compose logs

  # Show logs from a specific service:
  open-responses compose logs api

  # Follow log output:
  open-responses compose logs -f

  # Show last 10 lines of logs:
  open-responses compose logs --tail=10`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("logs", args)
	},
}

// PS command for listing containers
// This is a direct proxy to `docker compose ps` with argument passthrough
var psCmd = &cobra.Command{
	Use:   "ps [flags] [SERVICE...]",
	Short: "List containers (docker compose ps)",
	Long: `List containers.

This command is a direct proxy to 'docker compose ps' and accepts all the same flags.
Common flags include:
  -a, --all             Show all stopped containers (including those created by the run command)
  -q, --quiet           Only display IDs
  --format string       Format output using a custom template

Examples:
  # List all running containers:
  open-responses compose ps

  # List all containers, including stopped ones:
  open-responses compose ps -a

  # List only container IDs:
  open-responses compose ps -q`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("ps", args)
	},
}

// Build command for building or rebuilding services
// This is a direct proxy to `docker compose build` with argument passthrough
var buildCmd = &cobra.Command{
	Use:   "build [flags] [SERVICE...]",
	Short: "Build or rebuild services (docker compose build)",
	Long: `Build or rebuild services.

This command is a direct proxy to 'docker compose build' and accepts all the same flags.
Common flags include:
  --no-cache                 Do not use cache when building the image
  --pull                     Always attempt to pull a newer version of the image
  --progress string          Set type of progress output (auto, tty, plain, quiet) (default "auto")

Examples:
  # Build all services:
  open-responses compose build

  # Build specific services:
  open-responses compose build api ui

  # Build without using cache:
  open-responses compose build --no-cache`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("build", args)
	},
}

// Restart command for restarting containers
// This is a direct proxy to `docker compose restart` with argument passthrough
var restartCmd = &cobra.Command{
	Use:   "restart [flags] [SERVICE...]",
	Short: "Restart containers (docker compose restart)",
	Long: `Restart containers.

This command is a direct proxy to 'docker compose restart' and accepts all the same flags.
Common flags include:
  -t, --timeout int   Specify a shutdown timeout in seconds (default 10)

Examples:
  # Restart all services:
  open-responses compose restart

  # Restart specific services:
  open-responses compose restart api

  # Restart with a custom timeout:
  open-responses compose restart --timeout 30`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("restart", args)
	},
}

// Pull command for pulling service images
// This is a direct proxy to `docker compose pull` with argument passthrough
var pullCmd = &cobra.Command{
	Use:   "pull [flags] [SERVICE...]",
	Short: "Pull service images (docker compose pull)",
	Long: `Pull service images.

This command is a direct proxy to 'docker compose pull' and accepts all the same flags.
Common flags include:
  --ignore-pull-failures  Pull what it can and ignores images with pull failures
  --include-deps          Also pull services declared as dependencies
  --quiet                 Pull without printing progress information

Examples:
  # Pull all service images:
  open-responses compose pull

  # Pull specific service images:
  open-responses compose pull api db

  # Pull quietly:
  open-responses compose pull --quiet`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("pull", args)
	},
}

// CLAUDE-note-interact-cmds: Container interaction commands - Exec, Run for executing code in containers (~67 lines)
// Exec command for executing a command in a running container
// This is a direct proxy to `docker compose exec` with argument passthrough
var execCmd = &cobra.Command{
	Use:   "exec [flags] SERVICE COMMAND [ARGS...]",
	Short: "Execute a command in a running container (docker compose exec)",
	Long: `Execute a command in a running container.

This command is a direct proxy to 'docker compose exec' and accepts all the same flags.
Common flags include:
  -d, --detach          Run command in the background
  -e, --env strings     Set environment variables
  -w, --workdir string  Working directory inside the container

Examples:
  # Run an interactive shell in the api service container:
  open-responses compose exec api sh

  # Run a command in a service container:
  open-responses compose exec db psql -U postgres -d responses

  # Run a command with environment variables:
  open-responses compose exec -e VAR1=value1 api python script.py`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: SERVICE and COMMAND are required arguments.")
			fmt.Println("Usage: open-responses compose exec [flags] SERVICE COMMAND [ARGS...]")
			os.Exit(1)
		}
		executeDockerComposeWithArgs("exec", args)
	},
}

// Run command for running a one-off command in a service container
// This is a direct proxy to `docker compose run` with argument passthrough
var runCmd = &cobra.Command{
	Use:   "run [flags] SERVICE COMMAND [ARGS...]",
	Short: "Run a one-off command in a service container (docker compose run)",
	Long: `Run a one-off command in a service container.

This command is a direct proxy to 'docker compose run' and accepts all the same flags.
Common flags include:
  -d, --detach                Run container in background
  --rm                        Remove container after run (default: true)
  -e, --env strings           Set environment variables
  -v, --volume strings        Bind mount a volume
  -w, --workdir string        Working directory inside the container

Examples:
  # Run a one-off command in a service:
  open-responses compose run api python manage.py migrate

  # Run an interactive shell in a new container:
  open-responses compose run --rm api sh

  # Run with custom environment variables:
  open-responses compose run -e DEBUG=1 api python script.py`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: SERVICE and COMMAND are required arguments.")
			fmt.Println("Usage: open-responses compose run [flags] SERVICE COMMAND [ARGS...]")
			os.Exit(1)
		}
		executeDockerComposeWithArgs("run", args)
	},
}

// CLAUDE-note-config-cmds: Configuration management commands - Config for inspecting compose setup (~72 lines)
// Config command for displaying Docker Compose configuration
// This is a direct proxy to `docker compose config` with argument passthrough
var configCmd = &cobra.Command{
	Use:   "config [flags]",
	Short: "Validate and view the Docker Compose configuration (docker compose config)",
	Long: `Validate and view the Docker Compose configuration.

This command is a direct proxy to 'docker compose config' and accepts all the same flags.
Common flags include:
  --resolve-image-digests   Pin image tags to digests
  -q, --quiet               Only validate the configuration, don't print anything
  --services                Print the service names, one per line
  --volumes                 Print the volume names, one per line
  --profiles                Print the profile names, one per line

Examples:
  # Validate and display the Docker Compose configuration:
  open-responses compose config

  # Only validate the Docker Compose configuration:
  open-responses compose config -q

  # List the services defined in the Docker Compose file:
  open-responses compose config --services`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("config", args)
	},
}

// Top command for displaying running processes in containers
// This is a direct proxy to `docker compose top` with argument passthrough
var topCmd = &cobra.Command{
	Use:   "top [SERVICE...]",
	Short: "Display the running processes in containers (docker compose top)",
	Long: `Display the running processes in containers.

This command is a direct proxy to 'docker compose top'. You can specify service names
to only show processes for specific services.

Examples:
  # Show processes for all services:
  open-responses compose top

  # Show processes for specific services:
  open-responses compose top api db`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("top", args)
	},
}

// Stats command for displaying resource usage statistics
// This is a direct proxy to `docker compose stats` with argument passthrough
var statsCmd = &cobra.Command{
	Use:   "stats [SERVICE...]",
	Short: "Display a live stream of container resource usage (docker compose stats)",
	Long: `Display a live stream of container resource usage statistics.

This command is a direct proxy to 'docker compose stats' and shows CPU, memory, and network usage.
You can specify service names to only show statistics for specific services.

Examples:
  # Show stats for all services:
  open-responses compose stats

  # Show stats for specific services:
  open-responses compose stats api`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("stats", args)
	},
}

// CLAUDE-todo-cmd-groups: Consider grouping command registration by functionality categories (~18 lines)
// init registers all subcommands with the root command
func init() {
	// Add core commands directly to root
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(rootConfigCmd)

	// Add new user-friendly commands to root
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(rootStopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(rootLogsCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(keyCmd)

	// Add compose command to root
	rootCmd.AddCommand(composeCmd)

	// Add all Docker Compose related commands to the compose command group
	composeCmd.AddCommand(upCmd)
	composeCmd.AddCommand(stopCmd)
	composeCmd.AddCommand(downCmd)
	composeCmd.AddCommand(logsCmd)
	composeCmd.AddCommand(psCmd)
	composeCmd.AddCommand(buildCmd)
	composeCmd.AddCommand(restartCmd)
	composeCmd.AddCommand(pullCmd)
	composeCmd.AddCommand(execCmd)
	composeCmd.AddCommand(runCmd)
	composeCmd.AddCommand(configCmd)
	composeCmd.AddCommand(topCmd)
	composeCmd.AddCommand(statsCmd)
}

// main is the entry point for the CLI application
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// showConfig displays the current configuration settings
func showConfig() {
	// Check if config exists
	configExists, configPath := checkConfigExists()
	if !configExists {
		fmt.Println("No configuration found. Please run 'open-responses setup' first.")
		os.Exit(1)
	}

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse config
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	// Display config in a readable format
	fmt.Printf("Configuration file: %s\n\n", configPath)
	fmt.Printf("Version:          %s\n", config.Version)
	fmt.Printf("Created:          %s\n", valueOrDefault(config.CreatedAt, config.CreatedTime, "not set"))
	fmt.Printf("Last Updated:     %s\n", valueOrDefault(config.UpdatedAt, config.UpdatedTime, "not set"))
	fmt.Printf("Host:             %s\n", config.Host)
	fmt.Printf("Port:             %s\n", config.Port)
	fmt.Printf("Docker Tag:       %s\n", config.DockerTag)
	fmt.Printf("Base Compose URI: %s\n", config.BaseComposeURI)
	fmt.Printf("Env File:         %s\n", config.EnvFile)
	fmt.Printf("API Version:      %s\n", config.APIVersion)

	// Display environment variables
	fmt.Println("\nEnvironment Variables:")
	if len(config.Environment) == 0 {
		fmt.Println("  None defined")
	} else {
		// Show API keys with masking
		fmt.Println("  API Keys:")
		apiKeyCount := 0
		for key, value := range config.Environment {
			if strings.Contains(strings.ToLower(key), "api_key") || strings.Contains(strings.ToLower(key), "password") {
				apiKeyCount++
				fmt.Printf("    %s: %s\n", key, maskString(value))
			}
		}
		if apiKeyCount == 0 {
			fmt.Println("    None defined")
		}

		// Show other environment variables
		fmt.Println("  Other Settings:")
		otherCount := 0
		for key, value := range config.Environment {
			if !strings.Contains(strings.ToLower(key), "api_key") && !strings.Contains(strings.ToLower(key), "password") {
				otherCount++
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		if otherCount == 0 {
			fmt.Println("    None defined")
		}
	}

	// Display Docker Compose information
	if fileExists("responses-compose.yaml") {
		fmt.Println("\nDocker Compose:")
		fmt.Println("  responses-compose.yaml: Present")
	} else {
		fmt.Println("\nDocker Compose:")
		fmt.Println("  responses-compose.yaml: Not found (Run 'setup' to create it)")
	}
}

// valueOrDefault returns the first non-empty string or default value
func valueOrDefault(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return "not set"
}

// maskString masks a string for display (e.g., API keys)
func maskString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 8 {
		return "****"
	}
	// Show first 4 and last 4 chars
	return s[:4] + "..." + s[len(s)-4:]
}

// Global variable to hold the configuration
var config Config

// CLAUDE-note-setup-config: Configuration creation and management - Central setup function (~112 lines)
// setupConfig collects user input and creates or updates configuration files
func setupConfig() {
	// Check if config already exists
	configExists, configPath := checkConfigExists()

	currentTime := time.Now().Format(time.RFC3339)

	if configExists {
		// Read existing config
		configData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Error reading existing config: %v\n", err)
			os.Exit(1)
		}

		// Try to parse existing config
		err = json.Unmarshal(configData, &config)
		if err != nil {
			fmt.Printf("Error parsing existing config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Updating existing configuration...")

		// Update timestamps - support both formats
		config.UpdatedAt = currentTime
		config.UpdatedTime = currentTime

		// If timestamps don't exist, set creation time too
		if config.CreatedAt == "" && config.CreatedTime == "" {
			config.CreatedAt = currentTime
			config.CreatedTime = currentTime
		}
	} else {
		// Initialize new config with defaults
		config = Config{
			Version:        version,
			CreatedAt:      currentTime,
			UpdatedAt:      currentTime,
			CreatedTime:    currentTime,
			UpdatedTime:    currentTime,
			Host:           "127.0.0.1",
			Port:           "8080",
			DockerTag:      "latest_responses",
			BaseComposeURI: "default", // Use our built-in default template
			APIVersion:     "0.0.1",
			Environment:    make(map[string]string),
		}

		// Set default env file path
		gitRoot, err := findGitRootDir()
		if err == nil {
			config.EnvFile = filepath.Join(gitRoot, ".env")
		} else {
			// If git root not found, use current directory
			config.EnvFile = ".env"
		}

		fmt.Println("Creating new configuration...")
	}

	// Collect configuration values interactively
	err := survey.AskOne(&survey.Input{
		Message: "Host address for the API",
		Default: config.Host,
	}, &config.Host)
	if err != nil {
		fmt.Printf("Using default for Host: %s\n", config.Host)
	}

	err = survey.AskOne(&survey.Input{
		Message: "Port for the API",
		Default: config.Port,
	}, &config.Port)
	if err != nil {
		fmt.Printf("Using default for Port: %s\n", config.Port)
	}

	err = survey.AskOne(&survey.Input{
		Message: "Docker image tag",
		Default: config.DockerTag,
	}, &config.DockerTag)
	if err != nil {
		fmt.Printf("Using default for Docker tag: %s\n", config.DockerTag)
	}

	err = survey.AskOne(&survey.Input{
		Message: "Base Docker Compose file URL/path",
		Default: config.BaseComposeURI,
	}, &config.BaseComposeURI)
	if err != nil {
		fmt.Printf("Using default for Base Compose URI: %s\n", config.BaseComposeURI)
	}

	err = survey.AskOne(&survey.Input{
		Message: "Path to environment file",
		Default: config.EnvFile,
	}, &config.EnvFile)
	if err != nil {
		fmt.Printf("Using default for Env file: %s\n", config.EnvFile)
	}

	err = survey.AskOne(&survey.Input{
		Message: "API version",
		Default: config.APIVersion,
	}, &config.APIVersion)
	if err != nil {
		fmt.Printf("Using default for API version: %s\n", config.APIVersion)
	}

	// CLAUDE-note-env-vars: Environment variable definitions with validation and defaults (~80 lines)
	// Define the environment variables to collect
	envVars := []EnvVar{
		// Required variables that need to be set interactively
		{Name: "RESPONSES_API_KEY", Description: "Responses API Key", Default: "resp_" + generateRandomKey(48), Required: true},
		{Name: "MEMORY_STORE_PASSWORD", Description: "Memory Store Password", Default: generateRandomKey(24), Required: true},
		{Name: "OPENAI_API_KEY", Description: "OpenAI API Key (needed for embeddings and file parsing)", Default: "", Required: true},
		{Name: "UNSTRUCTURED_API_KEY", Description: "Unstructured API Key (for file parsing)", Default: "", Required: true},
		{Name: "BRAVE_API_KEY", Description: "Brave Search API Key (for web search tool)", Default: "", Required: true},

		// API configuration variables with sane defaults
		{Name: "RESPONSES_API_PORT", Description: "Port for the Responses API", Default: "8080", Required: false},
		{Name: "NODE_ENV", Description: "Node.js environment", Default: "production", Required: false},
		{Name: "LOG_LEVEL", Description: "Logging level", Default: "info", Required: false},
		{Name: "REQUEST_TIMEOUT", Description: "API request timeout in ms", Default: "120000", Required: false},
		{Name: "MAX_PAYLOAD_SIZE", Description: "Maximum request payload size", Default: "10mb", Required: false},
		{Name: "RATE_LIMIT_WINDOW", Description: "Rate limit window in ms", Default: "60000", Required: false},
		{Name: "RATE_LIMIT_MAX", Description: "Maximum requests per rate limit window", Default: "100", Required: false},

		// Optional provider API keys - mentioned but not requested interactively
		{Name: "ANTHROPIC_API_KEY", Description: "Anthropic API Key", Default: "", Required: false},
		{Name: "GEMINI_API_KEY", Description: "Gemini API Key", Default: "", Required: false},
		{Name: "OPENROUTER_API_KEY", Description: "OpenRouter API Key", Default: "", Required: false},
		{Name: "VOYAGE_API_KEY", Description: "Voyage API Key", Default: "", Required: false},
		{Name: "GROQ_API_KEY", Description: "Groq API Key", Default: "", Required: false},
		{Name: "CEREBRAS_API_KEY", Description: "Cerebras API Key", Default: "", Required: false},
		{Name: "CLOUDFLARE_API_KEY", Description: "Cloudflare API Key", Default: "", Required: false},
		{Name: "CLOUDFLARE_ACCOUNT_ID", Description: "Cloudflare Account ID", Default: "", Required: false},
		{Name: "GITHUB_API_KEY", Description: "GitHub API Key", Default: "", Required: false},
		{Name: "NVIDIA_NIM_API_KEY", Description: "NVIDIA NIM API Key", Default: "", Required: false},
		{Name: "LLAMA_API_KEY", Description: "Llama API Key", Default: "", Required: false},
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Description: "Google Application Credentials", Default: "", Required: false},
	}

	// Collect values for each environment variable
	if config.Environment == nil {
		config.Environment = make(map[string]string)
	}

	// First, handle required variables interactively
	fmt.Println("\nPlease enter the following required settings:")

	for _, env := range envVars {
		if !env.Required {
			continue // Only process required variables in this loop
		}

		// Use existing value as default if available
		defaultValue := env.Default
		if existing, ok := config.Environment[env.Name]; ok && existing != "" {
			defaultValue = existing
		}

		prompt := &survey.Input{
			Message: fmt.Sprintf("%s (%s)", env.Description, env.Name),
			Default: defaultValue,
		}

		var value string
		err := survey.AskOne(prompt, &value)
		if err != nil {
			// If there's an error with the interactive prompt (like no TTY),
			// just use the default value
			fmt.Printf("Using default for %s: %s\n", env.Name, defaultValue)
			value = defaultValue
		}
		config.Environment[env.Name] = value
	}

	// Copy over non-required variables with defaults
	for _, env := range envVars {
		if env.Required {
			continue // Skip required variables as they were handled above
		}

		// Only set if not already in environment
		if _, exists := config.Environment[env.Name]; !exists && env.Default != "" {
			config.Environment[env.Name] = env.Default
		}
	}

	// Inform the user about optional variables
	fmt.Println("\nConfiguration completed. Additional optional provider API keys can be set")
	fmt.Printf("by editing the environment file at: %s\n", config.EnvFile)

	// Create the .env file
	createEnvFile(config.Environment)

	// Create the responses-compose.yaml file
	createDockerComposeFile()

	// Check if we are in a git repository and if .env isn't already in .gitignore,
	// add it to ensure it doesn't get committed
	if _, err := findGitRootDir(); err == nil {
		// We are in a git repository, check if .env is in .gitignore
		addEnvToGitignore(filepath.Base(config.EnvFile))
	}

	// Save the config file
	saveConfigFile(config)

	fmt.Println("Configuration completed successfully!")
}

// addEnvToGitignore adds the environment file to .gitignore if not already present
// Cross-platform implementation using Go's standard library
func addEnvToGitignore(envFileName string) {
	gitRoot, err := findGitRootDir()
	if err != nil {
		return // Not a git repo, nothing to do
	}

	gitignorePath := filepath.Join(gitRoot, ".gitignore")

	// Check if .gitignore exists
	if !fileExists(gitignorePath) {
		// Create .gitignore if it doesn't exist
		err := os.WriteFile(gitignorePath, []byte(envFileName+"\n"), 0644)
		if err != nil {
			fmt.Printf("Warning: Failed to create .gitignore file: %v\n", err)
		}
		return
	}

	// Read existing .gitignore
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		fmt.Printf("Warning: Failed to read .gitignore file: %v\n", err)
		return
	}

	// Check if env file is already in .gitignore
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == envFileName {
			return // Already in .gitignore
		}
	}

	// Create updated content with new entry
	updatedContent := string(content)
	// Ensure we have a trailing newline before adding new content
	if !strings.HasSuffix(updatedContent, "\n") {
		updatedContent += "\n"
	}
	updatedContent += "\n# Environment variables\n" + envFileName + "\n"

	// Write the updated file
	err = os.WriteFile(gitignorePath, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to update .gitignore file: %v\n", err)
		return
	}

	fmt.Printf("Added %s to .gitignore\n", envFileName)
}

// CLAUDE-note-env-file: Environment file generation with structured sections (~110 lines)
// createEnvFile generates a .env file with the provided environment variables
func createEnvFile(envValues map[string]string) {
	content := "# Open Responses API Configuration File\n\n"

	// Add standard configuration variables
	content += "## Basic Configuration ##\n"
	content += fmt.Sprintf("HOST=%s\n", config.Host)
	content += fmt.Sprintf("PORT=%s\n", config.Port)
	content += fmt.Sprintf("DOCKER_TAG=%s\n", config.DockerTag)
	content += fmt.Sprintf("API_VERSION=%s\n", config.APIVersion)

	// Add required variables first with comments
	content += "\n## Authentication & Security ##\n"
	addEnvVar(&content, envValues, "RESPONSES_API_KEY")
	addEnvVar(&content, envValues, "MEMORY_STORE_PASSWORD")

	content += "\n## Required External API Keys ##\n"
	content += "# These keys are required for core functionality\n"
	addEnvVar(&content, envValues, "OPENAI_API_KEY")
	addEnvVar(&content, envValues, "UNSTRUCTURED_API_KEY")
	addEnvVar(&content, envValues, "BRAVE_API_KEY")

	// Add API configuration variables
	content += "\n## API Configuration ##\n"
	content += "# These settings control API behavior\n"
	addEnvVar(&content, envValues, "RESPONSES_API_PORT")
	addEnvVar(&content, envValues, "NODE_ENV")
	addEnvVar(&content, envValues, "LOG_LEVEL")
	addEnvVar(&content, envValues, "REQUEST_TIMEOUT")
	addEnvVar(&content, envValues, "MAX_PAYLOAD_SIZE")

	// Add rate limiting configuration
	content += "\n## Rate Limiting ##\n"
	content += "# Controls request rate limiting\n"
	addEnvVar(&content, envValues, "RATE_LIMIT_WINDOW")
	addEnvVar(&content, envValues, "RATE_LIMIT_MAX")

	// Add optional provider keys in an organized section
	content += "\n## Optional Provider API Keys ##\n"
	content += "# Enable one or more of these providers as needed\n"

	// Organize optional provider keys by category
	aiProviders := []string{
		"ANTHROPIC_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY",
		"VOYAGE_API_KEY", "GROQ_API_KEY", "CEREBRAS_API_KEY", "LLAMA_API_KEY",
	}

	cloudProviders := []string{
		"CLOUDFLARE_API_KEY", "CLOUDFLARE_ACCOUNT_ID", "NVIDIA_NIM_API_KEY",
	}

	serviceProviders := []string{
		"GITHUB_API_KEY", "GOOGLE_APPLICATION_CREDENTIALS",
	}

	// Add AI providers
	content += "\n# AI Model Providers\n"
	for _, name := range aiProviders {
		if value, ok := envValues[name]; ok {
			if value == "" {
				content += fmt.Sprintf("#%s=\"YOUR_%s\"\n", name, name)
			} else {
				content += fmt.Sprintf("%s=\"%s\"\n", name, value)
			}
		}
	}

	// Add cloud providers
	content += "\n# Cloud Service Providers\n"
	for _, name := range cloudProviders {
		if value, ok := envValues[name]; ok {
			if value == "" {
				content += fmt.Sprintf("#%s=\"YOUR_%s\"\n", name, name)
			} else {
				content += fmt.Sprintf("%s=\"%s\"\n", name, value)
			}
		}
	}

	// Add service providers
	content += "\n# Additional Service Providers\n"
	for _, name := range serviceProviders {
		if value, ok := envValues[name]; ok {
			if value == "" {
				content += fmt.Sprintf("#%s=\"YOUR_%s\"\n", name, name)
			} else {
				content += fmt.Sprintf("%s=\"%s\"\n", name, value)
			}
		}
	}

	// Ensure the directory exists
	envDir := filepath.Dir(config.EnvFile)
	if envDir != "." && envDir != "" {
		err := os.MkdirAll(envDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory for .env file: %v\n", err)
			os.Exit(1)
		}
	}

	err := os.WriteFile(config.EnvFile, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error creating .env file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created .env file at %s\n", config.EnvFile)
}

// CLAUDE-note-compose-template: Docker Compose template handling with fallback mechanisms (~80 lines)
// createDockerComposeFile generates a responses-compose.yaml file for the service
// If a responses-compose.yaml file exists locally, it uses that directly
// Otherwise, if BaseComposeURI is a URL, it downloads the file
// If it's a local path, it copies the file
func createDockerComposeFile() {
	var composeContent []byte
	var err error

	// First, check if a local responses-compose.yaml file exists
	if fileExists("responses-compose.yaml") {
		fmt.Println("Using existing responses-compose.yaml file")
		return
	}

	// Use default template when specified
	if config.BaseComposeURI == "default" {
		fmt.Println("Using default Docker Compose template")
		composeContent = []byte(getDefaultComposeTemplate())
	} else if strings.HasPrefix(config.BaseComposeURI, "http://") || strings.HasPrefix(config.BaseComposeURI, "https://") {
		fmt.Printf("Downloading Docker Compose template from %s\n", config.BaseComposeURI)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Make HTTP request
		resp, err := client.Get(config.BaseComposeURI)
		if err != nil {
			fmt.Printf("Error downloading Docker Compose template: %v\n", err)
			fmt.Println("Using default Docker Compose template instead.")
			composeContent = []byte(getDefaultComposeTemplate())
			// Update the config to use default template for future runs
			config.BaseComposeURI = "default"
		} else {
			defer resp.Body.Close()

			// Check if request was successful
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error downloading Docker Compose template: HTTP %d\n", resp.StatusCode)
				fmt.Println("Using default Docker Compose template instead.")
				composeContent = []byte(getDefaultComposeTemplate())
				// Update the config to use default template for future runs
				config.BaseComposeURI = "default"
			} else {
				// Read the response body
				composeContent, err = io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading Docker Compose template: %v\n", err)
					fmt.Println("Using default Docker Compose template instead.")
					composeContent = []byte(getDefaultComposeTemplate())
					// Update the config to use default template for future runs
					config.BaseComposeURI = "default"
				}
			}
		}
	} else if fileExists(config.BaseComposeURI) {
		// If it's a local file, read it
		fmt.Printf("Using local Docker Compose template from %s\n", config.BaseComposeURI)
		composeContent, err = os.ReadFile(config.BaseComposeURI)
		if err != nil {
			fmt.Printf("Error reading Docker Compose template: %v\n", err)
			fmt.Println("Using default Docker Compose template instead.")
			composeContent = []byte(getDefaultComposeTemplate())
			// Update the config to use default template for future runs
			config.BaseComposeURI = "default"
		}
	} else {
		// If it's neither a URL nor a local file, use the default template
		fmt.Printf("Docker Compose template not found at %s\n", config.BaseComposeURI)
		fmt.Println("Using default Docker Compose template instead.")
		composeContent = []byte(getDefaultComposeTemplate())
		// Update the config to use default template for future runs
		config.BaseComposeURI = "default"
	}

	// Write only to responses-compose.yaml (not docker-compose.yml)
	err = os.WriteFile("responses-compose.yaml", composeContent, 0644)
	if err != nil {
		fmt.Printf("Error creating responses-compose.yaml file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created responses-compose.yaml file")
}

// getDefaultComposeTemplate returns the default Docker Compose template
func getDefaultComposeTemplate() string {
	return `##########################
### OPEN RESPONSES API ###
##########################

# An open-source self-hosted version of the new [OpenAI Responses API](https://community.openai.com/t/introducing-the-responses-api/1140929).

name: open-responses

## ENVIRONMENT VARS ##
# (Recommended way to set them is using .env file instead of here)

x--shared-environment: &shared-environment
  ### REQUIRED ###
  # Responses API Key
  AGENTS_API_KEY: ${RESPONSES_API_KEY:?RESPONSES_API_KEY is required}
  AGENTS_API_KEY_HEADER_NAME: ${RESPONSES_API_KEY_HEADER_NAME:-Authorization}

  # TODO: Make these required after enabling doc search
  # OpenAI & Unstructured API key (needed for embeddings and file parsing)
  OPENAI_API_KEY: ${OPENAI_API_KEY}
  UNSTRUCTURED_API_KEY: ${UNSTRUCTURED_API_KEY}

  # Brave Search API Key (For web search tool)
  BRAVE_API_KEY: ${BRAVE_API_KEY}

  ### OPTIONAL ###
  # Service settings
  RESPONSES_API_PORT: ${RESPONSES_API_PORT:-8080}
  MEMORY_STORE_PASSWORD: ${MEMORY_STORE_PASSWORD:-obviously_not_a_safe_password}

  # Provider API Keys (Enable one or more as needed)
  ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
  GEMINI_API_KEY: ${GEMINI_API_KEY}
  OPENROUTER_API_KEY: ${OPENROUTER_API_KEY}
  VOYAGE_API_KEY: ${VOYAGE_API_KEY}
  GROQ_API_KEY: ${GROQ_API_KEY}
  CEREBRAS_API_KEY: ${CEREBRAS_API_KEY}
  CLOUDFLARE_API_KEY: ${CLOUDFLARE_API_KEY}
  CLOUDFLARE_ACCOUNT_ID: ${CLOUDFLARE_ACCOUNT_ID}
  LLAMA_API_KEY: ${LLAMA_API_KEY}
  GITHUB_API_KEY: ${GITHUB_API_KEY}
  NVIDIA_NIM_API_KEY: ${NVIDIA_NIM_API_KEY}
  GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}

  ### INTERNAL ###
  # Change **ONLY** if you know what you're doing
  EMBEDDING_MODEL_ID: ${EMBEDDING_MODEL_ID:-openai/text-embedding-3-large}
  ENABLE_RESPONSES: "True"
  PG_DSN: postgres://postgres:${MEMORY_STORE_PASSWORD:-obviously_not_a_safe_password}@memory-store:5432/postgres
  INTEGRATION_SERVICE_URL: ${INTEGRATION_SERVICE_URL:-http://integrations:8000}

## SERVICES ##

services:
  api:
    image: julepai/agents-api:${TAG:-responses-latest}
    environment:
      <<: *shared-environment
    ports:
      - "${RESPONSES_API_PORT:-8080}:8080" # map host to container port

  integrations:
    image: julepai/integrations:${TAG:-responses-latest}
    environment:
      <<: *shared-environment

  memory-store:
    image: timescale/timescaledb-ha:pg17
    environment:
      POSTGRES_PASSWORD: ${MEMORY_STORE_PASSWORD:-obviously_not_a_safe_password}
    volumes:
      - memory_store_data:/home/postgres/pgdata/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  vectorizer-worker:
    image: timescale/pgai-vectorizer-worker:latest
    environment:
      - PGAI_VECTORIZER_WORKER_DB_URL=postgres://postgres:${MEMORY_STORE_PASSWORD:-obviously_not_a_safe_password}@memory-store:5432/postgres
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    command: [ "--poll-interval", "5s" ]
    depends_on:
      memory-store:
        condition: service_healthy

  migration:
    image: julepai/memory-store-migrations:latest
    command: [ "-path", "/migrations", "-database", "postgres://postgres:${MEMORY_STORE_PASSWORD:-obviously_not_a_safe_password}@memory-store:5432/postgres?sslmode=disable" , "up"]
    restart: "no"
    depends_on:
      memory-store:
        condition: service_healthy

volumes:
  memory_store_data:
`
}

// saveConfigFile saves the configuration to the open-responses.json file
func saveConfigFile(config Config) {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error creating config JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(configFileName, configJSON, 0644)
	if err != nil {
		fmt.Printf("Error saving config file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved %s file\n", configFileName)
}

// upService launches the Docker Compose services using the 'up' command
// This function:
// 1. Uses the shared executeDockerComposeWithArgs function for prerequisites and command execution
// 2. Shows status and URLs when running in detached mode
func upService(args []string) {
	// Get configuration path for informational output
	_, configPath := checkConfigExists()

	// Inform the user which configuration we're using
	fmt.Printf("Starting services with configuration from %s\n", configPath)

	// Execute Docker Compose "up" command with all args
	// The common helper function handles all prerequisites and error handling
	executeDockerComposeWithArgs("up", args)

	// Determine if we're running in detached mode (-d or --detach flag)
	// When running attached, the command will block until terminated
	// When running detached, the command returns immediately
	isDetached := false
	for _, arg := range args {
		if arg == "-d" || arg == "--detach" {
			isDetached = true
			break
		}
	}

	// Only show additional information if running in detached mode
	// In attached mode, the Docker Compose output is streamed to the console
	if isDetached {
		fmt.Println("Service started successfully!")

		// Show a list of running containers
		statusCmd := executeDockerComposeCommand("ps")
		statusCmd.Stdout = os.Stdout
		statusCmd.Stderr = os.Stderr
		statusCmd.Run()

		// Show the access URLs for the services
		fmt.Printf("\nServices are now running!\n")
		fmt.Printf("API available at: http://%s:%s\n", config.Host, config.Environment["RESPONSES_API_PORT"])
		fmt.Printf("Admin UI available at: http://%s:%s\n", config.Host, config.Port)
	}
}

// stopService shuts down the Docker Compose services using the 'down' command
// This function:
// 1. Uses the shared executeDockerComposeWithArgs function for prerequisites and command execution
// 2. Shows a success message after stopping services
// This is used by both the 'stop' and 'down' commands
func stopService() {
	// Get configuration path for informational output
	_, configPath := checkConfigExists()

	// Inform the user which configuration we're using
	fmt.Printf("Stopping services with configuration from %s\n", configPath)

	// Execute Docker Compose "down" command
	// The common helper function handles all prerequisites and error handling
	executeDockerComposeWithArgs("down", []string{})

	fmt.Println("Service stopped successfully!")
}

// fileExists checks if a file exists at the given path
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// checkConfigExists checks if the config file exists in the current directory,
// parent directory, or git root directory.
func checkConfigExists() (bool, string) {
	// Check current directory
	if fileExists(configFileName) {
		return true, configFileName
	}

	// Check parent directory
	parentDir := filepath.Join("..", configFileName)
	if fileExists(parentDir) {
		return true, parentDir
	}

	// Try to find git root directory
	gitRoot, err := findGitRootDir()
	if err == nil {
		gitRootConfig := filepath.Join(gitRoot, configFileName)
		if fileExists(gitRootConfig) {
			return true, gitRootConfig
		}
	}

	return false, ""
}

// findGitRootDir attempts to find the root directory of the git repository
// This version is cross-platform and doesn't rely on shell commands
func findGitRootDir() (string, error) {
	// First try using git command, which is more accurate if available
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		// Trim newline from the end
		path := string(output)
		if len(path) > 0 && path[len(path)-1] == '\n' {
			path = path[:len(path)-1]
		}
		return path, nil
	}

	// Fallback to searching for .git directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for .git
	for {
		// Check if .git exists in the current directory
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root directory without finding .git
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

// generateRandomKey generates a random key of the specified length
func generateRandomKey(length int) string {
	// Generate random bytes
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// If we can't generate random bytes, use timestamp as fallback
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		paddedTimestamp := timestamp + strings.Repeat("0", length-len(timestamp))
		return paddedTimestamp[:length]
	}

	// Convert to a hex string and return the specified length
	return hex.EncodeToString(b)[:length]
}

// addEnvVar adds an environment variable to the content string with proper formatting
func addEnvVar(content *string, envValues map[string]string, name string) {
	if value, ok := envValues[name]; ok && value != "" {
		*content += fmt.Sprintf("%s=\"%s\"\n", name, value)
	} else {
		*content += fmt.Sprintf("%s=\"\"\n", name)
	}
}

// checkDockerInstalled checks if Docker is installed and in the PATH
func checkDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// CLAUDE-note-compose-versions: Docker Compose version detection and compatibility checking (~144 lines)
// DockerComposeType defines which Docker Compose command format to use
// This enum type helps track which version of Docker Compose is installed (if any)
type DockerComposeType int

const (
	DockerComposeV2   DockerComposeType = iota // docker compose (v2) - integrated plugin in Docker CLI
	DockerComposeV1                            // docker-compose (v1) - standalone binary
	DockerComposeNone                          // not installed or not found in PATH
)

// checkDockerComposeInstalled checks if any version of Docker Compose is available on the system
// Returns true if either Docker Compose v1 or v2 is installed and accessible
func checkDockerComposeInstalled() bool {
	return getDockerComposeType() != DockerComposeNone
}

// getDockerComposeType detects which Docker Compose variant is available on the system
// Tries Docker Compose v2 first (preferred), then falls back to v1 if available
// Returns the detected Docker Compose type (V2, V1, or None)
func getDockerComposeType() DockerComposeType {
	// Try checking for docker compose (v2) first - the preferred version
	// V2 is integrated into Docker as a plugin and uses "docker compose" command format
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return DockerComposeV2
	}

	// Fallback to checking for docker-compose (v1) - the legacy standalone version
	// V1 uses the "docker-compose" command format
	_, err := exec.LookPath("docker-compose")
	if err == nil {
		return DockerComposeV1
	}

	// Neither version is installed or available in PATH
	return DockerComposeNone
}

// checkDockerComposeVersion validates if Docker Compose version meets minimum requirements
// Requires Docker Compose v2.21.0 or higher for full compatibility
// Returns true if the installed version meets requirements
func checkDockerComposeVersion() bool {
	composeType := getDockerComposeType()

	switch composeType {
	case DockerComposeV2:
		// For V2, run docker compose version with --short flag to get just the version number
		cmd := exec.Command("docker", "compose", "version", "--short")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		// Parse version string (format example: "v2.23.0" or "2.23.0")
		versionStr := strings.TrimSpace(string(output))

		// Remove the 'v' prefix if present
		if strings.HasPrefix(versionStr, "v") {
			versionStr = versionStr[1:]
		}

		// Split version into components (major.minor.patch)
		parts := strings.Split(versionStr, ".")
		if len(parts) < 3 {
			// Not enough parts, can't determine version
			return false
		}

		// Parse major version component
		major, err := parseInt(parts[0])
		if err != nil || major < 2 {
			// Major version must be at least 2
			return false
		}

		// For major version > 2, it's already newer than required
		if major > 2 {
			return true
		}

		// Parse minor version component
		minor, err := parseInt(parts[1])
		if err != nil || minor < 21 {
			// Minor version must be at least 21
			return false
		}

		// If minor is > 21, it's newer than required
		if minor > 21 {
			return true
		}

		// Parse patch version component
		patch, err := parseInt(parts[2])
		if err != nil {
			return false
		}

		// Check if version is at least 2.21.0
		return patch >= 0

	case DockerComposeV1:
		// For v1, get the version string
		cmd := exec.Command("docker-compose", "--version")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		// Parse version string (format example: "docker-compose version 1.29.2, build 5becea4c")
		versionStr := strings.TrimSpace(string(output))

		// docker-compose v1 has format like: docker-compose version 1.29.2, build 5becea4c
		parts := strings.Split(versionStr, " ")
		if len(parts) < 3 {
			return false
		}

		// We don't need to check the version details since all v1 versions are considered outdated
		// Just confirm it's a known docker-compose v1 format
		_ = parts // Use parts to avoid compiler warnings

		// V1 is always considered older than required 2.21.0 version
		// We could implement version checking for v1 but it's not necessary since
		// all v1 versions are considered outdated compared to v2
		return false

	default:
		// Docker Compose not installed
		return false
	}
}

// parseInt parses a string to an integer, returning an error if it's not an integer
// Used for parsing version number components
func parseInt(s string) (int, error) {
	// Remove any additional information after the version number
	// Examples: "2-alpine", "2+dev"
	s = strings.Split(s, "-")[0]
	s = strings.Split(s, "+")[0]

	return strconv.Atoi(s)
}

// CLAUDE-note-compose-exec: Docker Compose command execution abstraction layer (~24 lines)
// executeDockerComposeCommand creates a command to execute Docker Compose operations
// Automatically selects the appropriate command format based on installed version:
// - For V2: "docker compose -f responses-compose.yaml <command> [args...]"
// - For V1: "docker-compose -f responses-compose.yaml <command> [args...]"
// Returns a ready-to-execute Cmd struct that can be further customized if needed
func executeDockerComposeCommand(command string, args ...string) *exec.Cmd {
	composeType := getDockerComposeType()

	var cmd *exec.Cmd

	if composeType == DockerComposeV2 {
		// For Docker Compose v2: docker compose -f responses-compose.yaml <command> [args...]
		composeArgs := append([]string{"compose", "-f", "responses-compose.yaml", command}, args...)
		cmd = exec.Command("docker", composeArgs...)
	} else {
		// For Docker Compose v1: docker-compose -f responses-compose.yaml <command> [args...]
		composeArgs := append([]string{"-f", "responses-compose.yaml", command}, args...)
		cmd = exec.Command("docker-compose", composeArgs...)
	}

	return cmd
}

// CLAUDE-todo-validation: Consider extracting prerequisite checks to a separate validation function (~84 lines)
// executeDockerComposeWithArgs is a high-level wrapper for Docker Compose commands
// It handles all the common setup and error handling for Docker Compose proxy commands:
// 1. Checks for Docker and Docker Compose installation
// 2. Loads configuration
// 3. Sets environment variables
// 4. Executes the command with provided arguments
// 5. Handles any errors and provides user-friendly output
func executeDockerComposeWithArgs(command string, args []string) {
	// Check if Docker is installed - required for all operations
	if !checkDockerInstalled() {
		fmt.Println("Docker is not installed or not in PATH.")
		fmt.Println("Please install Docker: https://docs.docker.com/get-docker/")
		os.Exit(1)
	}

	// Check if Docker Compose is installed (either V1 or V2)
	composeType := getDockerComposeType()
	if composeType == DockerComposeNone {
		fmt.Println("Docker Compose is not installed or not in PATH.")
		fmt.Println("Please install Docker Compose: https://docs.docker.com/compose/install/")
		os.Exit(1)
	}

	// For Docker Compose V2, validate the version meets minimum requirements
	// Docker Compose V2 >= 2.21.0 is recommended for full compatibility
	if composeType == DockerComposeV2 && !checkDockerComposeVersion() {
		fmt.Println("Docker Compose version is older than required 2.21.0.")
		fmt.Println("Please update Docker Compose: https://docs.docker.com/compose/install/")
		os.Exit(1)
	}

	// Load configuration from the open-responses.json file
	// This validates that setup has been completed
	configExists, configPath := checkConfigExists()
	if !configExists {
		fmt.Println("Configuration not found. Run 'open-responses setup' first.")
		os.Exit(1)
	}

	// Read existing config file to memory
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON config into our Config struct
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	// Verify that the environment file specified in the config exists
	if !fileExists(config.EnvFile) {
		fmt.Printf("Environment file not found at %s. Run 'setup' to recreate it.\n", config.EnvFile)
		os.Exit(1)
	}

	// Verify that responses-compose.yaml exists in the current directory
	// This file is created during setup from a template
	if !fileExists("responses-compose.yaml") {
		fmt.Println("Docker Compose file not found. Run 'setup' to recreate it.")
		os.Exit(1)
	}

	// Set environment variable to tell Docker Compose where to find .env file
	// This ensures all environment variables are available to the services
	os.Setenv("COMPOSE_DOTENV_PATH", config.EnvFile)

	// Create the appropriate Docker Compose command with the provided arguments
	cmd := executeDockerComposeCommand(command, args...)

	// Wire up stdout/stderr to the console for interactive output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error executing command '%s': %v\n", command, err)
		os.Exit(1)
	}
}

// CLAUDE-note-user-friendly-implementations: Implementation of user-friendly command functions

// startService is a user-friendly wrapper for 'compose up -d'
// This function provides a streamlined interface to start services with optimal defaults
func startService() {
	// Get configuration path for informational output
	_, configPath := checkConfigExists()

	// Inform the user which configuration we're using
	fmt.Printf("Starting services with configuration from %s\n", configPath)

	// Execute Docker Compose "up" command with detached mode flag
	// We always run in detached mode for this user-friendly command
	fmt.Println("Starting Responses API services...")
	executeDockerComposeWithArgs("up", []string{"-d"})

	// Show a divider to separate the command output from the status
	fmt.Println("\n-------------------------------------------------------------")

	// Give containers a moment to initialize
	time.Sleep(2 * time.Second)

	// Show service status after starting
	showServiceStatus()

	// Show access information
	fmt.Printf("\n🚀 Responses API is now running!\n")
	fmt.Printf("API available at: http://%s:%s\n", config.Host, config.Environment["RESPONSES_API_PORT"])
	fmt.Printf("Admin UI available at: http://%s:%s\n", config.Host, config.Port)
}

// stopRootService is a user-friendly wrapper for 'compose down'
// This function provides a clean interface to stop services and perform cleanup
func stopRootService() {
	// Get configuration path for informational output
	_, configPath := checkConfigExists()

	// Inform the user which configuration we're using
	fmt.Printf("Stopping services with configuration from %s\n", configPath)

	// Execute Docker Compose "down" command without additional arguments
	fmt.Println("Stopping Responses API services...")
	executeDockerComposeWithArgs("down", []string{})

	fmt.Println("\n✅ All services have been stopped and cleaned up!")
}

// showServiceStatus provides a detailed status overview of all services
// This function enhances the basic Docker Compose ps output with additional information
func showServiceStatus() {
	// Get configuration for proper context
	_, configPath := checkConfigExists()
	_, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("📊 Responses API Service Status")
	fmt.Println("-------------------------------------------------------------")

	// Execute Docker Compose "ps" to get container list
	// We capture the output rather than streaming to console so we can parse it
	cmd := executeDockerComposeCommand("ps")
	psOutput, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking service status: %v\n", err)
		os.Exit(1)
	}

	// Check if no containers are running
	if len(psOutput) == 0 || (len(psOutput) > 0 && !strings.Contains(string(psOutput), "_")) {
		fmt.Println("No services are currently running.")
		fmt.Println("\nTo start services, run: open-responses start")
		return
	}

	// Display the status output
	fmt.Println(string(psOutput))

	// Additional health check for running services
	// Use Docker Compose ps --format to get specific information about health
	healthCmd := executeDockerComposeCommand("ps", "--format", "table {{.Name}}\t{{.Status}}\t{{.Health}}")
	healthOutput, _ := healthCmd.Output()

	// Parse health status
	lines := strings.Split(string(healthOutput), "\n")
	if len(lines) > 1 {
		// Find any unhealthy services
		unhealthyServices := []string{}
		for _, line := range lines[1:] { // Skip header row
			if strings.Contains(line, "unhealthy") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					unhealthyServices = append(unhealthyServices, parts[0])
				}
			}
		}

		// Alert user about unhealthy services
		if len(unhealthyServices) > 0 {
			fmt.Println("\n⚠️ Warning: Some services are reporting unhealthy status:")
			for _, service := range unhealthyServices {
				fmt.Printf("  - %s\n", service)
			}
			fmt.Println("\nCheck logs for details: open-responses logs")
		} else {
			fmt.Println("\n✅ All services are healthy!")
		}
	}

	// Show resource usage if possible (Docker Compose stats)
	fmt.Println("\n💻 Resource Usage:")
	statsCmd := executeDockerComposeCommand("stats", "--no-stream")
	statsCmd.Stdout = os.Stdout
	statsCmd.Run()

	// Show access URLs if services are running
	fmt.Println("\n🔗 Access Information:")
	fmt.Printf("API available at: http://%s:%s\n", config.Host, config.Environment["RESPONSES_API_PORT"])
	fmt.Printf("Admin UI available at: http://%s:%s\n", config.Host, config.Port)
}

// showServiceLogs displays logs from services with optimal defaults
// This provides a more user-friendly interface compared to raw docker compose logs
func showServiceLogs(args []string) {
	// Add default flags for better user experience
	logArgs := []string{"-f", "--tail=100"}

	// Add any service arguments if specified
	if len(args) > 0 {
		logArgs = append(logArgs, args...)
	}

	fmt.Printf("Showing logs for %s services (Press Ctrl+C to exit)...\n\n",
		string(func() string {
			if len(args) > 0 {
				return strings.Join(args, ", ")
			}
			return "all"
		}()))

	// Execute logs command with enhanced default options
	executeDockerComposeWithArgs("logs", logArgs)
}

// initProject initializes a new project with best practices
// This is a streamlined version of setupConfig focused on new projects
func initProject() {
	// Check if configuration already exists
	configExists, configPath := checkConfigExists()

	if configExists {
		fmt.Printf("Configuration already exists at %s\n", configPath)
		fmt.Println("To update your configuration, run: open-responses setup")

		// Ask user if they want to continue
		var response string
		fmt.Print("Do you want to reinitialize your project? This will overwrite existing files. (y/N): ")
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Initialization cancelled.")
			os.Exit(0)
		}
	}

	// Create fresh project setup by delegating to the setup function
	fmt.Println("Initializing a new Responses API project...")

	// Create default workspace structure if needed
	for _, dir := range []string{"data", "config", "logs"} {
		if !fileExists(dir) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Error creating project directory '%s': %v\n", dir, err)
			} else {
				fmt.Printf("Created project directory: %s/\n", dir)
			}
		}
	}

	// Create README.md with helpful documentation if it doesn't exist
	if !fileExists("README.md") {
		readmeContent := `# Responses API Project

This project contains a self-hosted alternative to OpenAI's Responses API.

## Getting Started

1. Configure the service:
   ` + "```" + `
   open-responses setup
   ` + "```" + `

2. Start the services:
   ` + "```" + `
   open-responses start
   ` + "```" + `

3. View service status:
   ` + "```" + `
   open-responses status
   ` + "```" + `

4. View service logs:
   ` + "```" + `
   open-responses logs
   ` + "```" + `

5. Stop the services:
   ` + "```" + `
   open-responses stop
   ` + "```" + `

## Configuration

The service is configured using the ` + "`open-responses.json`" + ` file and environment variables.

To view or modify configuration:
` + "```" + `
open-responses config
` + "```" + `

## API Keys

To manage API keys:
` + "```" + `
open-responses key list
open-responses key generate
open-responses key set openai
` + "```" + `

## Advanced Commands

For advanced Docker Compose operations, use the compose command group:
` + "```" + `
open-responses compose ...
` + "```" + `
`

		err := os.WriteFile("README.md", []byte(readmeContent), 0644)
		if err != nil {
			fmt.Printf("Error creating README.md: %v\n", err)
		} else {
			fmt.Println("Created README.md with helpful documentation")
		}
	}

	// Run the standard setup process
	setupConfig()

	fmt.Println("\n🎉 Project initialization complete!")
	fmt.Println("Next steps:")
	fmt.Println("1. Start the services: open-responses start")
	fmt.Println("2. View service status: open-responses status")
	fmt.Println("3. View service logs: open-responses logs")
}

// updateComponents updates various components of the Responses API
func updateComponents() {
	// Check if configuration exists
	configExists, configPath := checkConfigExists()
	if !configExists {
		fmt.Println("No configuration found. Please run 'open-responses setup' first")
		os.Exit(1)
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	// Backup the current configuration
	backupPath := configPath + ".backup"
	err = os.WriteFile(backupPath, configData, 0644)
	if err != nil {
		fmt.Printf("Warning: Could not create backup of configuration: %v\n", err)
	} else {
		fmt.Printf("Configuration backed up to %s\n", backupPath)
	}

	// Update the Docker Compose file from template
	fmt.Println("Updating Docker Compose configuration...")
	createDockerComposeFile()

	// Pull latest Docker images
	fmt.Println("Pulling latest Docker images...")
	executeDockerComposeWithArgs("pull", []string{})

	// Update API versions if needed
	// In a full implementation, this would check for available updates
	// and apply migrations when necessary

	// Update timestamp in configuration
	config.UpdatedAt = time.Now().Format(time.RFC3339)
	config.UpdatedTime = config.UpdatedAt

	// Save updated configuration
	updatedConfigJSON, _ := json.MarshalIndent(config, "", "  ")
	err = os.WriteFile(configPath, updatedConfigJSON, 0644)
	if err != nil {
		fmt.Printf("Error saving updated configuration: %v\n", err)
	} else {
		fmt.Printf("Configuration updated at %s\n", configPath)
	}

	fmt.Println("\n✅ Update complete!")
	fmt.Println("To apply changes, restart your services:")
	fmt.Println("open-responses stop")
	fmt.Println("open-responses start")
}

// manageKeys provides functionality to manage API keys
func manageKeys(args []string) {
	// Check if configuration exists
	configExists, configPath := checkConfigExists()
	if !configExists {
		fmt.Println("No configuration found. Please run 'open-responses setup' first")
		os.Exit(1)
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	// Handle different key actions
	action := args[0]

	switch action {
	case "list":
		// List all API keys (masked)
		fmt.Println("🔑 API Keys:")
		fmt.Println("-------------------------------------------------------------")

		keyCount := 0
		for key, value := range config.Environment {
			if strings.Contains(strings.ToLower(key), "api_key") {
				keyCount++
				fmt.Printf("%-25s %s\n", key+":", maskString(value))
			}
		}

		if keyCount == 0 {
			fmt.Println("No API keys found in configuration.")
		}

	case "generate":
		// Generate a new API key
		var keyType string
		if len(args) > 1 {
			keyType = args[1]
		} else {
			keyType = "responses"
		}

		var keyPrefix string
		switch strings.ToLower(keyType) {
		case "responses":
			keyPrefix = "resp_"
		case "openai":
			keyPrefix = "sk-"
		case "anthropic":
			keyPrefix = "sk-ant-"
		case "unstructured":
			keyPrefix = "unst_"
		case "brave":
			keyPrefix = "brv_"
		default:
			keyPrefix = ""
		}

		// Generate appropriate length based on key type
		length := 32
		if strings.ToLower(keyType) == "responses" {
			length = 48
		}

		// Generate the key
		newKey := keyPrefix + generateRandomKey(length)

		fmt.Printf("Generated %s API key: %s\n", keyType, newKey)
		fmt.Println("\nTo use this key, set it in your configuration:")
		fmt.Printf("open-responses key set %s %s\n", strings.ToLower(keyType), newKey)

	case "set":
		// Set/update an API key
		if len(args) < 2 {
			fmt.Println("Error: Missing key type. Usage: open-responses key set <type> [value]")
			fmt.Println("Available key types: responses, openai, anthropic, unstructured, brave, etc.")
			os.Exit(1)
		}

		keyType := strings.ToUpper(args[1])
		var envName string

		// Map key type to environment variable name
		switch strings.ToLower(args[1]) {
		case "responses":
			envName = "RESPONSES_API_KEY"
		case "openai":
			envName = "OPENAI_API_KEY"
		case "anthropic":
			envName = "ANTHROPIC_API_KEY"
		case "unstructured":
			envName = "UNSTRUCTURED_API_KEY"
		case "brave":
			envName = "BRAVE_API_KEY"
		default:
			envName = keyType + "_API_KEY"
		}

		var keyValue string
		if len(args) > 2 {
			// Key value provided as argument
			keyValue = args[2]
		} else {
			// Prompt for key value
			fmt.Printf("Enter the %s API key: ", keyType)
			fmt.Scanln(&keyValue)
		}

		// Update the key in config
		if config.Environment == nil {
			config.Environment = make(map[string]string)
		}
		config.Environment[envName] = keyValue

		// Update env file
		content := "# Updated " + envName + "\n"
		envFile, err := os.ReadFile(config.EnvFile)
		if err == nil {
			lines := strings.Split(string(envFile), "\n")
			keyFound := false

			for i, line := range lines {
				if strings.HasPrefix(line, envName+"=") {
					lines[i] = envName + "=\"" + keyValue + "\""
					keyFound = true
					break
				}
			}

			if !keyFound {
				lines = append(lines, envName+"=\""+keyValue+"\"")
			}

			content = strings.Join(lines, "\n")
		} else {
			content += envName + "=\"" + keyValue + "\"\n"
		}

		err = os.WriteFile(config.EnvFile, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Warning: Could not update environment file: %v\n", err)
		}

		// Save updated config
		updatedConfigJSON, _ := json.MarshalIndent(config, "", "  ")
		err = os.WriteFile(configPath, updatedConfigJSON, 0644)
		if err != nil {
			fmt.Printf("Error saving updated configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ %s API key updated successfully!\n", keyType)
		fmt.Println("To apply changes, restart your services:")
		fmt.Println("open-responses stop")
		fmt.Println("open-responses start")

	default:
		fmt.Printf("Unknown key action: %s\n", action)
		fmt.Println("Available actions: list, generate, set")
		os.Exit(1)
	}
}

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

// EnvVar represents an environment variable with metadata for setup prompts
type EnvVar struct {
	Name        string // Environment variable name
	Description string // Human-readable description
	Default     string // Default value
	Required    bool   // Whether this env var is required during setup
}

// Config represents the CLI configuration stored in JSON format
type Config struct {
	Version        string            `json:"version"`        // CLI version
	CreatedAt      string            `json:"createdAt,omitempty"`      // Timestamp when config was created
	UpdatedAt      string            `json:"updatedAt,omitempty"`      // Timestamp when config was last updated
	CreatedTime    string            `json:"created_at,omitempty"`     // Alternate timestamp field for creation
	UpdatedTime    string            `json:"updated_at,omitempty"`     // Alternate timestamp field for update
	Host           string            `json:"host"`           // API host address
	Port           string            `json:"port"`           // API port
	DockerTag      string            `json:"docker_tag"`     // Docker image tag
	BaseComposeURI string            `json:"base_compose_uri"` // URL or path to base compose file
	EnvFile        string            `json:"env_file"`       // Path to env file
	APIVersion     string            `json:"api_version"`    // API version in semver format
	Environment    map[string]string `json:"environment"`    // Environment variable values
}

const (
	configFileName = "open-responses.json" // Name of the configuration file
	version        = "0.1.1"               // Current CLI version
)

// Root command for the CLI application
var rootCmd = &cobra.Command{
	Use:   "open-responses",
	Short: "A CLI for setting up an OpenAI Responses API alternative",
	Long:  `This CLI helps you set up and manage a Docker Compose service that mimics OpenAI's Responses API.`,
	
	// PersistentPreRunE runs before any command execution
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config check for the setup command
		if cmd.Use == "setup" {
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

// Setup command for initializing the service configuration
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up the Responses API service configuration",
	Run: func(cmd *cobra.Command, args []string) {
		setupConfig()
	},
}

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
  open-responses up
  
  # Start in detached mode (background):
  open-responses up -d
  
  # Force rebuild of all services:
  open-responses up --build --force-recreate`,
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
For complete cleanup, prefer the 'down' command.`,
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
defined in docker-compose.yml.

This command is a direct proxy to 'docker compose down'. It stops and removes
containers, networks, and optionally volumes created by 'up'.

Common flags include:
  -v, --volumes           Remove named volumes declared in the volumes section
  --rmi string            Remove images. Type must be one of: 'all': all images, 'local': only locally built images
  --remove-orphans        Remove containers for services not defined in the Compose file

Examples:
  # Stop and remove containers and networks:
  open-responses down
  
  # Stop and remove containers, networks, and volumes:
  open-responses down -v
  
  # Remove all images used by services:
  open-responses down --rmi all`,
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
  open-responses logs

  # Show logs from a specific service:
  open-responses logs api

  # Follow log output:
  open-responses logs -f

  # Show last 10 lines of logs:
  open-responses logs --tail=10`,
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
  open-responses ps

  # List all containers, including stopped ones:
  open-responses ps -a

  # List only container IDs:
  open-responses ps -q`,
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
  open-responses build

  # Build specific services:
  open-responses build api ui

  # Build without using cache:
  open-responses build --no-cache`,
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
  open-responses restart

  # Restart specific services:
  open-responses restart api

  # Restart with a custom timeout:
  open-responses restart --timeout 30`,
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
  open-responses pull

  # Pull specific service images:
  open-responses pull api db

  # Pull quietly:
  open-responses pull --quiet`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("pull", args)
	},
}

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
  open-responses exec api sh

  # Run a command in a service container:
  open-responses exec db psql -U postgres -d responses

  # Run a command with environment variables:
  open-responses exec -e VAR1=value1 api python script.py`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: SERVICE and COMMAND are required arguments.")
			fmt.Println("Usage: open-responses exec [flags] SERVICE COMMAND [ARGS...]")
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
  open-responses run api python manage.py migrate

  # Run an interactive shell in a new container:
  open-responses run --rm api sh

  # Run with custom environment variables:
  open-responses run -e DEBUG=1 api python script.py`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: SERVICE and COMMAND are required arguments.")
			fmt.Println("Usage: open-responses run [flags] SERVICE COMMAND [ARGS...]")
			os.Exit(1)
		}
		executeDockerComposeWithArgs("run", args)
	},
}

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
  open-responses config

  # Only validate the Docker Compose configuration:
  open-responses config -q

  # List the services defined in the Docker Compose file:
  open-responses config --services`,
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
  open-responses top

  # Show processes for specific services:
  open-responses top api db`,
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
  open-responses stats

  # Show stats for specific services:
  open-responses stats api`,
	Run: func(cmd *cobra.Command, args []string) {
		executeDockerComposeWithArgs("stats", args)
	},
}

// init registers all subcommands with the root command
func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(statsCmd)
}

// main is the entry point for the CLI application
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Global variable to hold the configuration
var config Config

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

	// Create the docker-compose.yml file
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
	
	// Add env file to .gitignore
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to open .gitignore file: %v\n", err)
		return
	}
	defer file.Close()
	
	if _, err := file.WriteString("\n# Environment variables\n" + envFileName + "\n"); err != nil {
		fmt.Printf("Warning: Failed to update .gitignore file: %v\n", err)
	}
	
	fmt.Printf("Added %s to .gitignore\n", envFileName)
}

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

// createDockerComposeFile generates a docker-compose.yml file for the service
// If BaseComposeURI is a URL, it downloads the file
// If it's a local path, it copies the file
func createDockerComposeFile() {
	var composeContent []byte
	var err error
	
	// Use default template when specified
	if config.BaseComposeURI == "default" {
		fmt.Println("Using default Docker Compose template")
		composeContent = []byte(getDefaultComposeTemplate())
	// Check if the BaseComposeURI is a URL
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
	
	// Write the compose file
	err = os.WriteFile("docker-compose.yml", composeContent, 0644)
	if err != nil {
		fmt.Printf("Error creating docker-compose.yml file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created docker-compose.yml file")
}

// getDefaultComposeTemplate returns the default Docker Compose template
func getDefaultComposeTemplate() string {
	return `version: '3.8'

services:
  api:
    image: julep/open-responses-api:${DOCKER_TAG:-latest_responses}
    ports:
      - "${HOST:-127.0.0.1}:${RESPONSES_API_PORT:-8080}:8080"
    environment:
      # API Configuration
      - RESPONSES_API_KEY=${RESPONSES_API_KEY}
      - API_VERSION=${API_VERSION:-0.0.1}
      - NODE_ENV=${NODE_ENV:-production}
      - LOG_LEVEL=${LOG_LEVEL:-info}
      - REQUEST_TIMEOUT=${REQUEST_TIMEOUT:-120000}
      - MAX_PAYLOAD_SIZE=${MAX_PAYLOAD_SIZE:-10mb}
      - RATE_LIMIT_WINDOW=${RATE_LIMIT_WINDOW:-60000}
      - RATE_LIMIT_MAX=${RATE_LIMIT_MAX:-100}
      # Storage Configuration
      - MEMORY_STORE_PASSWORD=${MEMORY_STORE_PASSWORD}
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${MEMORY_STORE_PASSWORD}
      - DB_NAME=responses
      - REDIS_HOST=memory
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${MEMORY_STORE_PASSWORD}
      # LLM Provider API Keys
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - UNSTRUCTURED_API_KEY=${UNSTRUCTURED_API_KEY}
      - BRAVE_API_KEY=${BRAVE_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
      - VOYAGE_API_KEY=${VOYAGE_API_KEY}
      - GROQ_API_KEY=${GROQ_API_KEY}
      - CEREBRAS_API_KEY=${CEREBRAS_API_KEY}
      - CLOUDFLARE_API_KEY=${CLOUDFLARE_API_KEY}
      - CLOUDFLARE_ACCOUNT_ID=${CLOUDFLARE_ACCOUNT_ID}
      - GITHUB_API_KEY=${GITHUB_API_KEY}
      - NVIDIA_NIM_API_KEY=${NVIDIA_NIM_API_KEY}
      - LLAMA_API_KEY=${LLAMA_API_KEY}
      - GOOGLE_APPLICATION_CREDENTIALS=${GOOGLE_APPLICATION_CREDENTIALS}
    restart: unless-stopped
    volumes:
      - api_data:/app/data
      - shared_data:/app/shared
    depends_on:
      db:
        condition: service_healthy
      memory:
        condition: service_healthy
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 2G
        reservations:
          cpus: '0.25'
          memory: 512M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 15s
    labels:
      - "com.julep.service=responses-api"
      - "com.julep.description=OpenAI Responses API Alternative"

  db:
    image: postgres:14-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${MEMORY_STORE_PASSWORD}
      - POSTGRES_DB=responses
      - PGDATA=/var/lib/postgresql/data/pgdata
    volumes:
      - db_data:/var/lib/postgresql/data
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.1'
          memory: 256M
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    command: ["postgres", "-c", "max_connections=100", "-c", "shared_buffers=256MB"]
    labels:
      - "com.julep.service=responses-db"
      - "com.julep.description=Database for Responses API"

  memory:
    image: redis:alpine
    command: redis-server --requirepass ${MEMORY_STORE_PASSWORD} --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 768M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${MEMORY_STORE_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s
    labels:
      - "com.julep.service=responses-memory"
      - "com.julep.description=Memory Store for Responses API"

  ui:
    image: julep/open-responses-ui:${DOCKER_TAG:-latest_responses}
    ports:
      - "${HOST:-127.0.0.1}:${PORT:-8080}:80"
    environment:
      - API_URL=http://${HOST:-127.0.0.1}:${RESPONSES_API_PORT:-8080}
      - RESPONSES_API_KEY=${RESPONSES_API_KEY}
      - NODE_ENV=${NODE_ENV:-production}
    depends_on:
      api:
        condition: service_healthy
    restart: unless-stopped
    volumes:
      - shared_data:/app/shared
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:80/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    labels:
      - "com.julep.service=responses-ui"
      - "com.julep.description=Management UI for Responses API"

volumes:
  db_data:
    labels:
      com.julep.usage: "postgres-data"
  api_data:
    labels:
      com.julep.usage: "api-data"
  redis_data:
    labels:
      com.julep.usage: "redis-data"
  shared_data:
    labels:
      com.julep.usage: "shared-data"
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
func findGitRootDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	// Trim newline from the end
	path := string(output)
	if len(path) > 0 && path[len(path)-1] == '\n' {
		path = path[:len(path)-1]
	}
	
	return path, nil
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

// executeDockerComposeCommand creates a command to execute Docker Compose operations
// Automatically selects the appropriate command format based on installed version:
// - For V2: "docker compose <command> [args...]"
// - For V1: "docker-compose <command> [args...]"
// Returns a ready-to-execute Cmd struct that can be further customized if needed
func executeDockerComposeCommand(command string, args ...string) *exec.Cmd {
	composeType := getDockerComposeType()
	
	var cmd *exec.Cmd
	
	if composeType == DockerComposeV2 {
		// For Docker Compose v2: docker compose <command> [args...]
		composeArgs := append([]string{"compose", command}, args...)
		cmd = exec.Command("docker", composeArgs...)
	} else {
		// For Docker Compose v1: docker-compose <command> [args...]
		composeArgs := append([]string{command}, args...)
		cmd = exec.Command("docker-compose", composeArgs...)
	}
	
	return cmd
}

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

	// Verify that docker-compose.yml exists in the current directory
	// This file is created during setup from a template
	if !fileExists("docker-compose.yml") {
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
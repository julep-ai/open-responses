# CLAUDE.md - Guidelines for Agentic Coding in this Repository

<!-- CLAUDE-note-overview: Project overview and purpose (~5 lines) -->

## Project Overview

- **Purpose**: CLI tool for setting up a self-hosted alternative to OpenAI's Responses API
- **Key Components**: Docker Compose service with API server, database, and management UI
- **API Compatibility**: Implements compatible endpoints with OpenAI's Responses API
- **Configuration Management**: Uses open-responses.json to track setup state and configuration

<!-- CLAUDE-note-build: Build commands and installation (~5 lines) -->

## Build Commands

- `npm run build` - Build for current platform
- `npm run build:all` - Build for all platforms (Linux, macOS, Windows)
- `pip install -e .` - Install Python package in development mode
- `python -m build` - Build Python package

<!-- CLAUDE-note-code-style: Code style guidelines for Go, Python, and JavaScript (~15 lines) -->

## Code Style Guidelines

- **Go**: Follow standard Go conventions

  - Use Cobra for CLI commands
  - Handle errors with detailed messages and proper exit codes
  - Group functions by functionality
  - Naming: camelCase for variables, PascalCase for exported functions

- **Python**: Follow PEP 8 conventions

  - Use docstrings for function documentation
  - Handle exceptions with try/except blocks
  - Use absolute imports

- **JavaScript**: Use modern JS syntax
  - Handle errors with try/catch blocks
  - Use proper process exit codes

<!-- CLAUDE-note-structure: Project structure and dependencies (~15 lines) -->

## Project Structure

- Hybrid Go/Python/Node.js project
- Go core with platform-specific binary distribution
- Python and Node.js wrappers for package distribution
- Docker Compose for running the API service components
- Cross-platform compatibility is essential

## Dependencies

- Go: github.com/spf13/cobra, github.com/AlecAivazis/survey/v2
- No external dependencies for Python/JavaScript wrappers

<!-- CLAUDE-note-config-system: Configuration system implementation details (~10 lines) -->

## Configuration System Implementation

- **Configuration File**: `open-responses.json` stores CLI configuration and environment variables
- **Setup Check**: All commands (except `setup`) verify configuration exists before running
- **Multi-location Search**: Configuration is searched for in:
  - Current directory
  - Parent directory
  - Git repository root directory
- **First-Run Behavior**: Fails with instructions to run `setup` if no configuration found
- **Setup Process**: Interactive prompts collect required settings and create config
- **Update Process**: When run with existing config, preserves previous values as defaults

<!-- CLAUDE-note-config-schema: Configuration schema with JSON example (~25 lines) -->

## Configuration Schema

```json
{
  "version": "0.1.0", // CLI version
  "createdAt": "2025-03-23T...", // ISO timestamp of creation (camelCase format)
  "updatedAt": "2025-03-23T...", // ISO timestamp of last update (camelCase format)
  "created_at": "2025-03-23T...", // ISO timestamp of creation (snake_case format)
  "updated_at": "2025-03-23T...", // ISO timestamp of last update (snake_case format)
  "host": "127.0.0.1", // API host address
  "port": "8080", // API port
  "docker_tag": "latest_responses", // Docker image tag to use
  "base_compose_uri": "https://u.julep.ai/responses-compose.yaml", // Template source
  "env_file": "/path/to/.env", // Path to .env file
  "api_version": "0.0.1", // API compatibility version
  "environment": {
    // Environment variables
    "API_PORT": "3000",
    "API_KEY": "sk-open-responses-default-key",
    "RESPONSE_TIMEOUT": "300",
    "DB_USER": "user",
    "DB_PASSWORD": "password",
    "DB_NAME": "responses_db"
  }
}
```

Note: The configuration supports both camelCase (`createdAt`, `updatedAt`) and snake_case (`created_at`, `updated_at`) timestamp formats for backward compatibility.

<!-- CLAUDE-note-env-organization: Environment variable organization with example (~45 lines) -->

## Environment Variable Organization

The `.env` file is organized into logical sections for better readability:

```bash
# Open Responses API Configuration File

## Basic Configuration ##
HOST=127.0.0.1
PORT=8080
DOCKER_TAG=latest_responses
API_VERSION=0.0.1

## Authentication & Security ##
RESPONSES_API_KEY="resp_abc123..."
MEMORY_STORE_PASSWORD="securepassword"

## Required External API Keys ##
# These keys are required for core functionality
OPENAI_API_KEY="sk-..."
UNSTRUCTURED_API_KEY="unst_..."
BRAVE_API_KEY="brv_..."

## API Configuration ##
# These settings control API behavior
RESPONSES_API_PORT="8080"
NODE_ENV="production"
LOG_LEVEL="info"
REQUEST_TIMEOUT="120000"
MAX_PAYLOAD_SIZE="10mb"

## Rate Limiting ##
# Controls request rate limiting
RATE_LIMIT_WINDOW="60000"
RATE_LIMIT_MAX="100"

## Optional Provider API Keys ##
# Enable one or more of these providers as needed

# AI Model Providers
ANTHROPIC_API_KEY="sk-ant-..."
#GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
#OPENROUTER_API_KEY="YOUR_OPENROUTER_API_KEY"
...
```

This organization provides:

1. Clear sections for different types of configuration
2. Helpful comments for users to understand settings
3. Logical grouping of related environment variables
4. Default values that work well for most deployments
5. Optional settings clearly marked with comments

<!-- CLAUDE-note-markers: Code navigation with marker comments system (~50 lines) -->

## Code Navigation with Marker Comments

The main.go file contains special marker comments to help with navigation and understanding key sections. These markers follow this format:

```
// CLAUDE-{type}-{descriptor}: Brief description (~XX lines)
```

Where:

- `{type}` can be `note` (for important sections to understand), `todo` (for improvement suggestions), or `warning` (for areas that need caution)
- `{descriptor}` is a descriptive identifier using kebab-case that summarizes the section (e.g., "data-structs", "setup-config")
- Each marker includes an approximate line count for the section (e.g., "~30 lines")

### Using Markers for Navigation

When working with main.go or other large files:

1. **First step when exploring a new file**: Search for `CLAUDE-` markers to get an overview of key areas
2. **Navigating complex code**: Jump between related markers to understand connected functionality
3. **Understanding the codebase structure**: Use markers to identify logical groupings

### Current Markers in main.go

```
CLAUDE-note-data-structs: Core data structures for environment configuration and CLI setup (~30 lines)
CLAUDE-note-root-cmd: Command structure definitions - Root command and CLI entrypoint (~23 lines)
CLAUDE-note-setup-cmd: Setup command - Primary configuration initialization workflow (~10 lines)
CLAUDE-note-lifecycle-cmds: Service lifecycle management commands - Up, Stop, Down for controlling services (~84 lines)
CLAUDE-note-monitor-cmds: Service monitoring commands - Logs, PS, Top, Stats for observing service state (~138 lines)
CLAUDE-note-interact-cmds: Container interaction commands - Exec, Run for executing code in containers (~67 lines)
CLAUDE-note-config-cmds: Configuration management commands - Config for inspecting compose setup (~72 lines)
CLAUDE-todo-cmd-groups: Consider grouping command registration by functionality categories (~18 lines)
CLAUDE-note-setup-config: Configuration creation and management - Central setup function (~112 lines)
CLAUDE-note-env-vars: Environment variable definitions with validation and defaults (~80 lines)
CLAUDE-note-env-file: Environment file generation with structured sections (~110 lines)
CLAUDE-note-compose-template: Docker Compose template handling with fallback mechanisms (~80 lines)
CLAUDE-note-compose-versions: Docker Compose version detection and compatibility checking (~144 lines)
CLAUDE-note-compose-exec: Docker Compose command execution abstraction layer (~24 lines)
CLAUDE-todo-validation: Consider extracting prerequisite checks to a separate validation function (~84 lines)
```

The approximate line counts help you determine how many lines to view after each marker, making it easier to read complete logical sections without unnecessary scrolling.

### Current Markers in CLAUDE.md

CLAUDE.md also contains marker comments for easier navigation:

```
<!-- CLAUDE-note-overview: Project overview and purpose (~5 lines) -->
<!-- CLAUDE-note-build: Build commands and installation (~5 lines) -->
<!-- CLAUDE-note-code-style: Code style guidelines for Go, Python, and JavaScript (~15 lines) -->
<!-- CLAUDE-note-structure: Project structure and dependencies (~15 lines) -->
<!-- CLAUDE-note-config-system: Configuration system implementation details (~10 lines) -->
<!-- CLAUDE-note-config-schema: Configuration schema with JSON example (~25 lines) -->
<!-- CLAUDE-note-compose-impl: Docker Compose implementation features (~15 lines) -->
<!-- CLAUDE-note-compose-config: Docker Compose configuration details with service examples (~85 lines) -->
<!-- CLAUDE-note-env-organization: Environment variable organization with example (~45 lines) -->
<!-- CLAUDE-note-markers: Code navigation with marker comments system (~50 lines) -->
<!-- CLAUDE-note-compose-compat: Docker Compose V1/V2 compatibility details (~65 lines) -->
```

### Adding New Markers

When modifying very long files or adding complex functionality:

1. Add appropriate marker comments for any new sections or major changes
2. Use consistent naming and numbering (increment from the highest existing number)
3. Keep descriptions brief (1-2 lines) but informative
4. Include approximate line count information for each section (~XX lines)
5. Prioritize marking code that's complex, important, or likely to need future maintenance
6. For any file over 500 lines, consider adding marker comments to help with navigation
7. When exploring large files, first use GrepTool to find markers (e.g., pattern "CLAUDE-")
8. When viewing a marked section, use the line count to set an appropriate limit parameter

<!-- CLAUDE-note-compose-compat: Docker Compose V1/V2 compatibility details (~65 lines) -->

## Docker Compose V1/V2 Compatibility

The `open-responses` CLI seamlessly handles both Docker Compose V1 (standalone binary) and V2 (Docker plugin) formats:

1. **Detection Strategy**:

   - First tries `docker compose version` to check for V2
   - Falls back to looking for `docker-compose` in PATH for V1
   - Displays appropriate error messages if neither is found

2. **Command Format Mapping**:

   - V2: `docker compose <command> [args...]`
   - V1: `docker-compose <command> [args...]`

3. **Command Abstraction**:

   - Uses `executeDockerComposeCommand()` helper to dynamically select correct format
   - Automatically passes through all command-line arguments

4. **Proxy Command Implementation**:

   - All common Docker Compose commands are available as direct proxies with argument passthrough:
     - `up`: Create and start containers
     - `down`: Stop and remove containers, networks, and optionally volumes
     - `logs`: View output from containers
     - `ps`: List containers
     - `build`: Build or rebuild services
     - `restart`: Restart service containers
     - `pull`: Pull service images
     - `exec`: Execute a command in a running container
     - `run`: Run a one-off command in a service container
     - `config`: Validate and view the Docker Compose configuration
     - `top`: Display the running processes in containers
     - `stats`: Display a live stream of container resource usage
   - Legacy `stop` command is kept for backward compatibility (alias for `down`)
   - Each command validates prerequisites, loads configuration, and passes through all flags and arguments

5. **Examples**:

   ```
   # User command:                    # Executed as:
   open-responses up -d               # docker compose up -d                  (V2)
                                      # docker-compose up -d                  (V1)

   open-responses down -v             # docker compose down -v                (V2)
                                      # docker-compose down -v                (V1)

   open-responses logs -f api         # docker compose logs -f api            (V2)
                                      # docker-compose logs -f api            (V1)

   open-responses exec db psql        # docker compose exec db psql           (V2)
                                      # docker-compose exec db psql           (V1)

   open-responses build --no-cache    # docker compose build --no-cache       (V2)
                                      # docker-compose build --no-cache       (V1)

   open-responses config --services   # docker compose config --services      (V2)
                                      # docker-compose config --services      (V1)

   open-responses top api             # docker compose top api                (V2)
                                      # docker-compose top api                (V1)

   open-responses stats               # docker compose stats                  (V2)
                                      # docker-compose stats                  (V1)
   ```

6. **User Experience Benefits**:
   - Users don't need to know which version they have installed
   - CLI provides appropriate installation instructions if needed
   - All Docker Compose flags and arguments are supported
   - Terminal output from Docker Compose is preserved

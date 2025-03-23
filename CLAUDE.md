# CLAUDE.md - Guidelines for Agentic Coding in this Repository

## Project Overview
- **Purpose**: CLI tool for setting up a self-hosted alternative to OpenAI's Responses API
- **Key Components**: Docker Compose service with API server, database, and management UI
- **API Compatibility**: Implements compatible endpoints with OpenAI's Responses API
- **Configuration Management**: Uses open-responses.json to track setup state and configuration

## Build Commands
- `npm run build` - Build for current platform
- `npm run build:all` - Build for all platforms (Linux, macOS, Windows)
- `pip install -e .` - Install Python package in development mode
- `python -m build` - Build Python package

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

## Project Structure
- Hybrid Go/Python/Node.js project
- Go core with platform-specific binary distribution
- Python and Node.js wrappers for package distribution
- Docker Compose for running the API service components
- Cross-platform compatibility is essential

## Dependencies
- Go: github.com/spf13/cobra, github.com/AlecAivazis/survey/v2
- No external dependencies for Python/JavaScript wrappers

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

## Configuration Schema
```json
{
  "version": "0.1.0",              // CLI version
  "createdAt": "2025-03-23T...",   // ISO timestamp of creation (camelCase format)
  "updatedAt": "2025-03-23T...",   // ISO timestamp of last update (camelCase format)
  "created_at": "2025-03-23T...",  // ISO timestamp of creation (snake_case format)
  "updated_at": "2025-03-23T...",  // ISO timestamp of last update (snake_case format)
  "host": "127.0.0.1",             // API host address
  "port": "8080",                  // API port 
  "docker_tag": "latest_responses", // Docker image tag to use
  "base_compose_uri": "https://u.julep.ai/responses-compose.yaml", // Template source
  "env_file": "/path/to/.env",     // Path to .env file
  "api_version": "0.0.1",          // API compatibility version
  "environment": {                 // Environment variables
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

## Docker Compose Implementation
- **Template Source**: Can be loaded from a URL or local file
- **Fallback Mechanism**: Uses built-in default template if specified source is unavailable
- **Environment Variables**: All settings are injected from config and .env file
- **Container Images**: Uses specified docker_tag for versioning
- **Command Compatibility**: Supports both Docker Compose V1 and V2 command formats
- **Version Check**: Verifies Docker Compose V2 is ≥ 2.21.0 for compatibility
- **Command Proxy**: Up and down commands pass through all arguments to Docker Compose
- **Version Detection**: Auto-detects installed Docker Compose version
- **Helpful Error Messages**: Provides informative error messages with installation instructions
- **Resource Management**: Configures appropriate CPU and memory limits for services
- **Health Checks**: Implements service health monitoring for proper startup sequencing
- **Performance Tuning**: Applies database and cache optimizations for improved reliability
- **Container Labels**: Adds metadata labels for easier container management

## Docker Compose Configuration Details

The default Docker Compose configuration includes the following optimizations:

### API Service
```yaml
api:
  # Resource Limits
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 2G
      reservations:
        cpus: '0.25'
        memory: 512M
  # Health Monitoring
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 15s
  # Configuration
  environment:
    # Rate Limiting
    - RATE_LIMIT_WINDOW=60000
    - RATE_LIMIT_MAX=100
    # Request Handling
    - REQUEST_TIMEOUT=120000
    - MAX_PAYLOAD_SIZE=10mb
    # Logging
    - LOG_LEVEL=info
    - NODE_ENV=production
```

### Database Service
```yaml
db:
  # Resource Limits
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        cpus: '0.1'
        memory: 256M
  # Health Monitoring
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U postgres"]
    interval: 10s
    timeout: 5s
    retries: 5
    start_period: 10s
  # Performance Tuning
  command: ["postgres", "-c", "max_connections=100", "-c", "shared_buffers=256MB"]
```

### Memory Service (Redis)
```yaml
memory:
  # Memory Management
  command: redis-server --requirepass ${MEMORY_STORE_PASSWORD} --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
  # Resource Limits
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 768M
      reservations:
        cpus: '0.1'
        memory: 128M
  # Health Monitoring
  healthcheck:
    test: ["CMD", "redis-cli", "-a", "${MEMORY_STORE_PASSWORD}", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5
    start_period: 5s
```

### Service Dependencies
The configuration implements proper service dependency management:
- The API service depends on both database and memory services
- Dependencies are defined with health checks to ensure services are truly ready
- The UI service depends on the API service to ensure proper startup sequence

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
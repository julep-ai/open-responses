# Open Responses CLI

A CLI tool for setting up a self-hosted alternative to OpenAI's Responses API. This API lets you create and manage open-ended AI responses for your applications, similar to OpenAI's Responses API, but fully under your control.

## Features

- Easy setup with Docker Compose
- Compatible API endpoints with OpenAI's Responses API
- Management UI for creating, viewing, and managing responses
- Local data storage with PostgreSQL
- Customizable authentication and timeout settings

## Installation

You can install this CLI using either npm or pip:

### Using npm

```bash
npx open-responses
```

Or install it globally:

```bash
npm install -g open-responses
open-responses
```

### Using pip

```bash
pipx open-responses
```

Or install it globally:

```bash
pip install open-responses
open-responses
```

## Usage

### First-time Setup

Before using any commands, you must run the setup command:

```bash
open-responses setup
```

This will:
- Ask for configuration settings with default values:
  - Host (default: 127.0.0.1)
  - Port (default: 8080)
  - Docker tag (default: latest_responses)
  - Base Docker Compose URI (default: https://u.julep.ai/responses-compose.yaml)
  - Environment file location (default: .env in Git root or current directory)
  - API version (default: 0.0.1)
- Ask for API configuration values (port, authentication key, timeout)
- Create a .env file with your settings
- Download or generate a docker-compose.yml file with the necessary services:
  - API server
  - Database
  - Management UI
- Create a configuration file (open-responses.json) to track your settings

The CLI will automatically check for this configuration before running any other commands. If the configuration file doesn't exist, it will prompt you to run the setup command first.

### Configuration File

The CLI stores its configuration in `open-responses.json`, which can be located in:
- The current directory
- The parent directory
- The Git repository root directory

The configuration file tracks:
- All user-defined settings
- Environment variable values
- Creation and update timestamps (both `camelCase` and `snake_case` formats are supported)
- File locations and version information

When you run `setup` again with an existing configuration, it will let you update your settings while preserving your previous values as defaults. If timestamps are missing from an existing configuration, they'll be added automatically when the configuration is updated.

### API Configuration

The API service includes the following configuration options with sensible defaults:

#### Basic Settings
- `HOST`: Host address for the API (default: `127.0.0.1`)
- `PORT`: Port for the UI service (default: `8080`)
- `RESPONSES_API_PORT`: Port for the API service (default: `8080`)
- `DOCKER_TAG`: Docker image tag (default: `latest_responses`)
- `API_VERSION`: API version (default: `0.0.1`)

#### Performance & Limits
- `NODE_ENV`: Node.js environment (default: `production`)
- `LOG_LEVEL`: Logging level (default: `info`)
- `REQUEST_TIMEOUT`: API request timeout in ms (default: `120000` - 2 minutes)
- `MAX_PAYLOAD_SIZE`: Maximum request payload size (default: `10mb`)
- `RATE_LIMIT_WINDOW`: Rate limit window in ms (default: `60000` - 1 minute)
- `RATE_LIMIT_MAX`: Maximum requests per rate limit window (default: `100`)

#### Resource Allocation
The Docker Compose configuration also includes resource limits to ensure stable operation:

- API Service: 1 CPU, 2GB memory (min: 0.25 CPU, 512MB)
- Database: 1 CPU, 1GB memory (min: 0.1 CPU, 256MB)
- Redis: 0.5 CPU, 768MB memory (min: 0.1 CPU, 128MB)
- UI: 0.5 CPU, 512MB memory (min: 0.1 CPU, 128MB)

These settings provide a good balance for most deployments, but you can adjust them in the `docker-compose.yml` file if needed.

### Starting the service

```bash
open-responses up [flags]
```

This command is a direct proxy to `docker compose up` and accepts all the same flags.

Common examples:
```bash
# Start in the foreground with output streaming to the console:
open-responses up

# Start in detached mode (background):
open-responses up -d

# Force rebuild of all services:
open-responses up --build --force-recreate
```

This command will:
- Verify that Docker and Docker Compose are installed and compatible
- Check that setup has been completed
- Start your Responses API service with any provided flags
- Show the status of running containers (in detached mode)
- Display access URLs (in detached mode)

### Stopping the service

```bash
open-responses down [flags]
```

This command is a direct proxy to `docker compose down` and accepts all the same flags.

Common examples:
```bash
# Stop and remove containers and networks:
open-responses down

# Stop and remove containers, networks, and volumes:
open-responses down -v
```

This command will:
- Verify that Docker and Docker Compose are installed
- Check that setup has been completed
- Stop and remove the Responses API service containers, networks, and optionally volumes

For backward compatibility, the `stop` command is also available, which performs the same function as `down`.

### Additional Docker Compose Commands

The CLI provides proxy commands to all common Docker Compose operations:

#### Viewing Logs

```bash
open-responses logs [flags] [SERVICE...]
```

Examples:
```bash
# Show logs from all services:
open-responses logs

# Show logs from a specific service:
open-responses logs api

# Follow log output:
open-responses logs -f

# Show last 10 lines of logs:
open-responses logs --tail=10
```

#### Listing Containers

```bash
open-responses ps [flags]
```

Examples:
```bash
# List all running containers:
open-responses ps

# List all containers, including stopped ones:
open-responses ps -a
```

#### Building Services

```bash
open-responses build [flags] [SERVICE...]
```

Examples:
```bash
# Build all services:
open-responses build

# Build specific services:
open-responses build api ui

# Build without using cache:
open-responses build --no-cache
```

#### Restarting Containers

```bash
open-responses restart [flags] [SERVICE...]
```

Examples:
```bash
# Restart all services:
open-responses restart

# Restart specific services:
open-responses restart api

# Restart with a custom timeout:
open-responses restart --timeout 30
```

#### Pulling Service Images

```bash
open-responses pull [flags] [SERVICE...]
```

Examples:
```bash
# Pull all service images:
open-responses pull

# Pull specific service images:
open-responses pull api db
```

#### Executing Commands in Containers

```bash
open-responses exec [flags] SERVICE COMMAND [ARGS...]
```

Examples:
```bash
# Run an interactive shell in the api service container:
open-responses exec api sh

# Run a command in a service container:
open-responses exec db psql -U postgres -d responses
```

#### Running One-off Commands

```bash
open-responses run [flags] SERVICE COMMAND [ARGS...]
```

Examples:
```bash
# Run a one-off command in a service:
open-responses run api python manage.py migrate

# Run an interactive shell in a new container:
open-responses run --rm api sh
```

#### Validating Configuration

```bash
open-responses config [flags]
```

Examples:
```bash
# Validate and display the Docker Compose configuration:
open-responses config

# Only validate the Docker Compose configuration:
open-responses config -q

# List the services defined in the Docker Compose file:
open-responses config --services
```

#### Viewing Container Processes

```bash
open-responses top [SERVICE...]
```

Examples:
```bash
# Show processes for all services:
open-responses top

# Show processes for specific services:
open-responses top api db
```

#### Monitoring Resource Usage

```bash
open-responses stats [SERVICE...]
```

Examples:
```bash
# Show real-time resource usage for all services:
open-responses stats

# Show resource usage for specific services:
open-responses stats api
```

## API Endpoints

Once your service is running, the following endpoints will be available:

- `POST /v1/responses` - Create a new response
- `GET /v1/responses/{id}` - Retrieve a response
- `GET /v1/responses` - List all responses
- `DELETE /v1/responses/{id}` - Delete a response

You can access the management UI at `http://localhost:8080` (or your configured port).

## Requirements

- Docker must be installed on your system
- Docker Compose must be installed (either as a standalone binary or integrated plugin)
  - Docker Compose V2 (≥ 2.21.0) is recommended for best compatibility
  - Docker Compose V1 is supported but with limited functionality
- No other runtime dependencies required (no Node.js or Python needed for running the service)

The CLI will check Docker and Docker Compose requirements and provide helpful instructions if they're not met.

## How it works

This CLI is built with Go and compiled to native binaries for Windows, macOS, and Linux.
When installed via npm or pip, the appropriate binary for your platform is used automatically.

The service itself runs in Docker containers, providing a compatible alternative to OpenAI's Responses API.

## Development

### Project Structure

- `main.go`: Core CLI functionality built with Go
- `open_responses/__init__.py`: Python wrapper for binary distribution
- `scripts/postinstall.js`: Node.js script for platform detection and setup
- `bin/`: Directory for compiled binaries

### Building from Source

Build for your current platform:
```bash
npm run build
```

Build for all platforms:
```bash
npm run build:all
```

This will generate binaries in the `bin/` directory:
- `bin/open-responses-linux`
- `bin/open-responses-macos`
- `bin/open-responses-win.exe`

### Installing for Development

For Python:
```bash
pip install -e .
```

For npm:
```bash
npm link
```

## License

Apache-2.0
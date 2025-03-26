# Contributing to Open Responses

Thank you for considering contributing to Open Responses! This document outlines the process for contributing to the project and the standards we follow.

## Code of Conduct

We expect all contributors to follow our code of conduct: be respectful, considerate, and collaborative.

## Setting Up Development Environment

1. Clone the repository:

   ```bash
   git clone https://github.com/julep-ai/open-responses.git
   cd open-responses
   ```

2. Install dependencies:

   ```bash
   # Install JavaScript dependencies
   npm install

   # Install Python dependencies
   uv pip install ruff

   # Install git hooks for automatic linting/formatting
   npm run install:hooks
   ```

3. Build the project:
   ```bash
   npm run build
   ```

## Development Workflow

1. Create a new branch for your feature or bugfix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following our coding guidelines below.

3. Run tests to ensure your changes work correctly:

   ```bash
   go test ./test/...
   ```

4. Commit your changes with a clear descriptive message.

5. Push your branch and create a pull request.

## Code Formatting and Linting

We use automated tools to maintain code quality. The git hooks you installed will automatically check your code before each commit, but you can also run these checks manually:

```bash
# Format all code
npm run format:all

# Lint all code
npm run lint:all

# Format/lint individual languages
npm run format       # JavaScript/JSON/Markdown files
npm run py:format    # Python files
npm run go:format    # Go files
npm run lint         # JavaScript files
npm run py:lint      # Python files
```

## Coding Guidelines

### Go

- Follow standard Go conventions
- Use Cobra for CLI commands
- Handle errors with detailed messages and proper exit codes
- Group functions by functionality
- Use camelCase for variables, PascalCase for exported functions

### Python

- Follow PEP 8 conventions
- Use docstrings for function documentation
- Handle exceptions with try/except blocks
- Use absolute imports

### JavaScript

- Use modern JS syntax
- Handle errors with try/catch blocks
- Use proper process exit codes

## Pull Request Process

1. Ensure all linting and formatting checks pass
2. Include tests for new functionality
3. Update documentation, including README.md if necessary
4. Ensure all tests pass
5. Request a review from at least one maintainer

## Building for All Platforms

Before submitting a PR that changes the build process or dependencies, ensure the project builds for all supported platforms:

```bash
npm run build:all
```

## License

By contributing to this project, you agree to license your contributions under the Apache-2.0 license.

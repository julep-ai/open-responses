#!/usr/bin/env bash

# Install git hooks script for open-responses

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
HOOKS_DIR="${REPO_ROOT}/.git/hooks"

# Ensure hooks directory exists
mkdir -p "${HOOKS_DIR}"

# Generate pre-commit hook content
echo "Generating pre-commit hook..."
cat > "${HOOKS_DIR}/pre-commit" << 'EOL'
#!/usr/bin/env bash

# Exit on error
set -e

echo "Running pre-commit hooks..."

# JavaScript/JSON/MD formatting and linting
echo "Running JavaScript/JSON/MD formatting and linting..."
if command -v npx &> /dev/null; then
  # Run only on staged files using lint-staged
  npx lint-staged
else
  echo "Warning: npx not found, skipping JavaScript/JSON/MD formatting and linting"
fi

# Python formatting and linting with Ruff
echo "Running Python formatting and linting with Ruff..."
if command -v uvx &> /dev/null; then
  uvx ruff format --exclude="**/[.]git/**" .
  uvx ruff check --exclude="**/[.]git/**" --fix .
else
  echo "Warning: uvx not found, skipping Python formatting and linting"
fi

# Go formatting
echo "Running Go formatting..."
if command -v go &> /dev/null; then
  go fmt ./...
else
  echo "Warning: go not found, skipping Go formatting"
fi

echo "Pre-commit hooks completed successfully!"
EOL

# Make hook executable
chmod +x "${HOOKS_DIR}/pre-commit"
echo "Pre-commit hook installed successfully."

# Install dev dependencies if needed
echo "Checking for dev dependencies..."
if [[ -f "${REPO_ROOT}/package.json" ]]; then
  echo "Installing JavaScript dev dependencies..."
  npm install --save-dev husky lint-staged prettier eslint
fi

if command -v uv &> /dev/null; then
  echo "Installing Python dev dependencies..."
  uv pip install ruff
  echo "Installation complete. You can now use 'uvx ruff' to run ruff commands."
else
  echo "Warning: uv not found, skipping Python dev dependencies installation"
  echo "Please install uv and run: uv pip install ruff"
fi

echo "Git hooks installed successfully!"
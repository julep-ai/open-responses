#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const os = require("os");

// Determine the platform-specific binary to use
const platform = os.platform();
let binaryName;
if (platform === "win32") {
  binaryName = "open-responses-win.exe";
} else if (platform === "darwin") {
  binaryName = "open-responses-macos";
} else if (platform === "linux") {
  binaryName = "open-responses-linux";
} else {
  console.error(`Unsupported platform: ${platform}`);
  process.exit(1);
}

// Paths for source and destination binaries
const sourcePath = path.join(__dirname, "..", "bin", binaryName);
const destPath = path.join(__dirname, "..", "bin", "open-responses");
const destPathWithExt = platform === "win32" ? `${destPath}.exe` : destPath;

// Check if bin directory exists, if not create it
const binDir = path.join(__dirname, "..", "bin");
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

try {
  // Check if source binary exists
  if (!fs.existsSync(sourcePath)) {
    console.warn(`Platform-specific binary not found: ${sourcePath}`);
    console.warn("You may need to build the binaries first with: npm run build:all");

    // Create shell script wrapper if missing
    if (!fs.existsSync(destPath)) {
      const shellScript = `#!/bin/sh
SCRIPT_DIR=$(dirname "$0")
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')

case "$PLATFORM" in
    linux*)     BINARY="open-responses-linux" ;;
    darwin*)    BINARY="open-responses-macos" ;;
    msys*|mingw*|cygwin*|windows*) BINARY="open-responses-win.exe" ;;
    *)          echo "Unsupported platform: $PLATFORM" >&2; exit 1 ;;
esac

exec "$SCRIPT_DIR/$BINARY" "$@"
`;
      fs.writeFileSync(destPath, shellScript);
      fs.chmodSync(destPath, 0o755);
      console.log("Created shell script wrapper");
    }

    // Exit gracefully
    process.exit(0);
  }

  // Copy the platform-specific binary to the generic name
  fs.copyFileSync(sourcePath, destPathWithExt);

  // Make binary executable (not needed on Windows)
  if (platform !== "win32") {
    fs.chmodSync(destPathWithExt, 0o755);
  }

  console.log(`Successfully installed open-responses CLI for ${platform}`);
} catch (err) {
  console.error(`Error during installation: ${err.message}`);
  process.exit(1);
}

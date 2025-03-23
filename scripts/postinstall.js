#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync } = require('child_process');

// Determine the platform-specific binary to use
const platform = os.platform();
let binaryName;
if (platform === 'win32') {
  binaryName = 'open-responses-win.exe';
} else if (platform === 'darwin') {
  binaryName = 'open-responses-macos';
} else if (platform === 'linux') {
  binaryName = 'open-responses-linux';
} else {
  console.error(`Unsupported platform: ${platform}`);
  process.exit(1);
}

// Paths for source and destination binaries
const sourcePath = path.join(__dirname, '..', 'bin', binaryName);
const destPath = path.join(__dirname, '..', 'bin', 'open-responses');
const destPathWithExt = platform === 'win32' ? `${destPath}.exe` : destPath;

try {
  // Copy the platform-specific binary to the generic name
  fs.copyFileSync(sourcePath, destPathWithExt);
  
  // Make binary executable (not needed on Windows)
  if (platform !== 'win32') {
    fs.chmodSync(destPathWithExt, 0o755);
  }
  
  console.log(`Successfully installed open-responses CLI for ${platform}`);
} catch (err) {
  console.error(`Error during installation: ${err.message}`);
  process.exit(1);
}
import os
import platform
import subprocess
import sys


def main():
    """Entry point for the CLI when installed via pip"""
    # Get the path to the binary based on the platform
    bin_dir = os.path.join(os.path.dirname(__file__), "bin")

    if platform.system() == "Windows":
        binary = os.path.join(bin_dir, "open-responses-win.exe")
    elif platform.system() == "Darwin":  # macOS
        binary = os.path.join(bin_dir, "open-responses-macos")
    elif platform.system() == "Linux":
        binary = os.path.join(bin_dir, "open-responses-linux")
    else:
        print(f"Unsupported platform: {platform.system()}")
        sys.exit(1)

    # Make sure the binary is executable
    if platform.system() != "Windows":
        os.chmod(binary, 0o755)

    # Execute the binary with the same arguments
    try:
        result = subprocess.run([binary] + sys.argv[1:])
        sys.exit(result.returncode)
    except Exception as e:
        print(f"Error executing the CLI binary: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()

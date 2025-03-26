#!/usr/bin/env python3
"""
Script to download the latest Responses API templates for the open-responses CLI.
This script downloads:
1. The latest Docker Compose template
2. The latest .env example file

Usage:
    python scripts/download_templates.py
"""

import os
import shutil
import sys
import urllib.request


def create_temp_dir():
    """Create a temporary directory for downloads if it doesn't exist."""
    temp_dir = os.path.join(os.getcwd(), "tmp")
    if not os.path.exists(temp_dir):
        os.makedirs(temp_dir)
    return temp_dir


def download_file(url, dest_path):
    """Download a file from a URL to the specified path."""
    try:
        print(f"Downloading {url}...")
        urllib.request.urlretrieve(url, dest_path)
        print(f"Successfully downloaded to {dest_path}")
        return True
    except Exception as e:
        print(f"Error downloading {url}: {str(e)}")
        return False


def update_main_go_docker_tag():
    """Update the default Docker tag in main.go from 'latest_responses' to 'responses-latest'."""
    main_go_path = os.path.join(os.getcwd(), "main.go")
    if not os.path.exists(main_go_path):
        print("Error: main.go not found!")
        return False

    with open(main_go_path) as f:
        content = f.read()

    # Update the Docker tag
    updated_content = content.replace(
        'DockerTag:      "latest_responses"', 'DockerTag:      "responses-latest"'
    )

    with open(main_go_path, "w") as f:
        f.write(updated_content)

    print("Updated default Docker tag in main.go")
    return True


def update_openai_api_key_required():
    """Update the OPENAI_API_KEY to be required during setup."""
    main_go_path = os.path.join(os.getcwd(), "main.go")
    if not os.path.exists(main_go_path):
        print("Error: main.go not found!")
        return False

    with open(main_go_path) as f:
        content = f.read()

    # Update the OPENAI_API_KEY required flag
    openai_required_pattern = (
        '{Name: "OPENAI_API_KEY", '
        'Description: "OpenAI API Key (needed for embeddings and file parsing)", '
        'Default: "", Required: true},'
    )
    if openai_required_pattern in content:
        print("OPENAI_API_KEY is already marked as required")
    else:
        old_pattern = (
            '{Name: "OPENAI_API_KEY", '
            'Description: "OpenAI API Key (needed for embeddings and file parsing)", '
            'Default: "", Required: false},'
        )
        new_pattern = (
            '{Name: "OPENAI_API_KEY", '
            'Description: "OpenAI API Key (needed for embeddings and file parsing)", '
            'Default: "", Required: true},'
        )
        updated_content = content.replace(old_pattern, new_pattern)

        with open(main_go_path, "w") as f:
            f.write(updated_content)

        print("Updated OPENAI_API_KEY to be required during setup")

    return True


def main():
    """Main function to download and update templates."""
    # Check for debug flag
    debug_mode = "--debug" in sys.argv
    print("Starting template download process...")

    if debug_mode:
        print("[DEBUG MODE] Running in debug mode - no files will be modified")

    # Create temp directory
    temp_dir = create_temp_dir()

    # Define file paths
    compose_url = "https://u.julep.ai/responses-compose.yaml"
    env_url = "https://u.julep.ai/responses-env.example"

    compose_temp_path = os.path.join(temp_dir, "responses-compose.yaml")
    env_temp_path = os.path.join(temp_dir, "responses-env.example")

    # Download files
    compose_success = download_file(compose_url, compose_temp_path)
    env_success = download_file(env_url, env_temp_path)

    if not compose_success or not env_success:
        print("Error: Failed to download one or more template files")
        sys.exit(1)

    # Update default Docker tag in main.go
    tag_updated = update_main_go_docker_tag()

    # Make OPENAI_API_KEY required
    openai_required_updated = update_openai_api_key_required()

    # Cleanup
    if debug_mode:
        print("[DEBUG MODE] Would clean up temporary files")
    else:
        print("Cleaning up temporary files...")
        shutil.rmtree(temp_dir)

    print("\n✅ Template update completed successfully!")
    print("The following changes were made:")
    print("1. Downloaded and processed the latest Docker Compose template")
    print("2. Downloaded and processed the latest environment file example")
    if tag_updated:
        print("3. Updated default Docker tag from 'latest_responses' to 'responses-latest'")
    if openai_required_updated:
        print("4. Updated OPENAI_API_KEY to be required during setup")


if __name__ == "__main__":
    main()

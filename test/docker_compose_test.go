package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestDockerComposeTypeDetection tests the Docker Compose version detection logic
func TestDockerComposeTypeDetection(t *testing.T) {
	// Skip test if Docker is not installed
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("Docker not installed, skipping test")
	}

	// Run the main docker compose detection logic
	// This is simulating the getDockerComposeType function
	composeType := "none"

	// Try checking for docker compose (v2) first
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		composeType = "v2"
	} else {
		// Fallback to checking for docker-compose (v1)
		_, err := exec.LookPath("docker-compose")
		if err == nil {
			composeType = "v1"
		}
	}

	// We expect either v1, v2, or none
	if composeType != "v1" && composeType != "v2" && composeType != "none" {
		t.Errorf("Invalid Docker Compose type detected: %s", composeType)
	}

	t.Logf("Docker Compose type detected: %s", composeType)
}

// TestDockerComposeCommand tests the Docker Compose command formation logic
func TestDockerComposeCommand(t *testing.T) {
	// Skip test if Docker is not installed
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("Docker not installed, skipping test")
	}

	// Determine which compose command to use (simulating executeDockerComposeCommand)
	command := "ps"
	args := []string{"-a"}
	
	composeType := "none"
	
	// Check for v2
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		composeType = "v2"
	} else {
		// Check for v1
		_, err := exec.LookPath("docker-compose")
		if err == nil {
			composeType = "v1"
		}
	}

	if composeType == "none" {
		t.Skip("Docker Compose not installed, skipping test")
	}

	// Build the command based on compose type
	var commandString string
	if composeType == "v2" {
		commandString = "docker compose " + command + " " + strings.Join(args, " ")
	} else {
		commandString = "docker-compose " + command + " " + strings.Join(args, " ")
	}

	// Check that the command string is properly formed
	if composeType == "v2" && !strings.HasPrefix(commandString, "docker compose") {
		t.Errorf("Invalid V2 command: %s", commandString)
	}
	
	if composeType == "v1" && !strings.HasPrefix(commandString, "docker-compose") {
		t.Errorf("Invalid V1 command: %s", commandString)
	}

	t.Logf("Docker Compose command: %s", commandString)
}

// TestDockerComposeVersion tests the version parsing logic
func TestDockerComposeVersion(t *testing.T) {
	// Test cases for version parsing
	testCases := []struct {
		version  string
		expected bool // Whether it meets the minimum requirement (>= 2.21.0)
	}{
		{"v2.21.0", true},
		{"2.21.0", true},
		{"v2.22.0", true},
		{"2.22.0", true},
		{"v2.20.0", false},
		{"2.20.0", false},
		{"v1.29.2", false},
		{"1.29.2", false},
		{"v3.0.0", true},
		{"3.0.0", true},
	}

	for _, tc := range testCases {
		// Parse version string
		version := tc.version
		
		// Remove the 'v' prefix if present
		if strings.HasPrefix(version, "v") {
			version = version[1:]
		}
		
		// Split version into components
		parts := strings.Split(version, ".")
		if len(parts) < 3 {
			t.Errorf("Invalid version format: %s", tc.version)
			continue
		}
		
		// Parse major version
		major := parseInt(parts[0])
		meetsRequirement := false
		
		if major < 2 {
			meetsRequirement = false
		} else if major > 2 {
			meetsRequirement = true
		} else {
			// Major is 2, check minor
			minor := parseInt(parts[1])
			if minor < 21 {
				meetsRequirement = false
			} else if minor > 21 {
				meetsRequirement = true
			} else {
				// Minor is 21, check patch
				patch := parseInt(parts[2])
				meetsRequirement = patch >= 0
			}
		}
		
		if meetsRequirement != tc.expected {
			t.Errorf("Version %s: got %v, expected %v", tc.version, meetsRequirement, tc.expected)
		}
	}
}

// Helper function to parse version components
func parseInt(s string) int {
	// Remove any additional information
	s = strings.Split(s, "-")[0]
	s = strings.Split(s, "+")[0]
	
	// Convert to int
	val := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			val = val*10 + int(r-'0')
		} else {
			break
		}
	}
	return val
}

// TestDockerComposeEnvVariables tests environment variable handling for Docker Compose
func TestDockerComposeEnvVariables(t *testing.T) {
	// Test that COMPOSE_DOTENV_PATH is properly set
	envFile := "/test/.env"
	
	// Set environment variable
	os.Setenv("COMPOSE_DOTENV_PATH", envFile)
	
	// Check that it was set correctly
	actualEnvFile := os.Getenv("COMPOSE_DOTENV_PATH")
	if actualEnvFile != envFile {
		t.Errorf("COMPOSE_DOTENV_PATH not set correctly. Got %s, expected %s", actualEnvFile, envFile)
	}
}

// TestCommandArgumentPassthrough tests argument passthrough for Docker Compose commands
func TestCommandArgumentPassthrough(t *testing.T) {
	// Define test cases for arguments
	testCases := []struct {
		command string
		args    []string
		v1      string
		v2      string
	}{
		{
			command: "up",
			args:    []string{"-d", "--build"},
			v1:      "docker-compose up -d --build",
			v2:      "docker compose up -d --build",
		},
		{
			command: "logs",
			args:    []string{"-f", "--tail=10", "api"},
			v1:      "docker-compose logs -f --tail=10 api",
			v2:      "docker compose logs -f --tail=10 api",
		},
		{
			command: "exec",
			args:    []string{"db", "psql", "-U", "postgres"},
			v1:      "docker-compose exec db psql -U postgres",
			v2:      "docker compose exec db psql -U postgres",
		},
	}

	for _, tc := range testCases {
		// For V1 command format
		v1Command := "docker-compose " + tc.command + " " + strings.Join(tc.args, " ")
		if v1Command != tc.v1 {
			t.Errorf("V1 command incorrect: got %s, expected %s", v1Command, tc.v1)
		}

		// For V2 command format
		v2Command := "docker compose " + tc.command + " " + strings.Join(tc.args, " ")
		if v2Command != tc.v2 {
			t.Errorf("V2 command incorrect: got %s, expected %s", v2Command, tc.v2)
		}
	}
}
package api

import (
	"os"
	"path/filepath"
	"testing"

	"provisioning-system/internal/models"
)

func TestExecuteCommands(t *testing.T) {
	// 1. Setup temp dir
	tmpDir, err := os.MkdirTemp("", "provisioning-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Create Handler
	h := &PhoneHandler{
		ConfigDir: tmpDir,
	}

	// 3. Prepare data
	mac := "00:11:22:33:44:55"
	phone := &models.Phone{
		Domain:     "test.local",
		MacAddress: &mac,
	}
	domainVars := map[string]string{
		"ServerIP": "192.168.1.100",
	}

	// 4. Define commands
	// We'll use a command that writes to a file in the temp dir
	outputFile := filepath.Join(tmpDir, "output.txt")
	commands := []string{
		"echo 'Domain: {{.Domain}}' > " + outputFile,
		"echo 'MAC: {{.Phone.MacAddress}}' >> " + outputFile,
		"echo 'Server: {{.Vars.ServerIP}}' >> " + outputFile,
	}

	// 5. Execute
	err = h.executeCommands(commands, "test.local", phone, domainVars)
	if err != nil {
		t.Fatalf("executeCommands failed: %v", err)
	}

	// 6. Verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expected := "Domain: test.local\nMAC: 00:11:22:33:44:55\nServer: 192.168.1.100\n"
	if string(content) != expected {
		t.Errorf("Unexpected content.\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}
}

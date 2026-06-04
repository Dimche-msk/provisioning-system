package provisioner

import (
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v6"
	"provisioning-system/internal/config"
	"provisioning-system/internal/models"
)

func TestDirectorySorting(t *testing.T) {
	// 1. Create a Manager instance
	m := NewManager(&config.SystemConfig{
		Domains: []config.DomainSettings{
			{
				Name: "test.local",
				Variables: map[string]string{
					"company": "Acme",
				},
			},
		},
	})

	// 2. Prepare mock phones
	mac1 := "001122334455"
	ph1 := models.Phone{
		Domain:      "test.local",
		Type:        "phone",
		MacAddress:  &mac1,
		PhoneNumber: func() *string { s := "102"; return &s }(),
		Description: "Alice Smith",
	}

	mac2 := "001122334466"
	ph2 := models.Phone{
		Domain:      "test.local",
		Type:        "gateway",
		MacAddress:  &mac2,
		PhoneNumber: func() *string { s := "192.168.1.10"; return &s }(),
		Description: "Gateway 1",
		Lines: []models.PhoneLine{
			{
				Type:           "Line",
				AccountNumber:  1,
				AdditionalInfo: `{"description":"Bob Jones","user_name":"103"}`,
			},
			{
				Type:           "Line",
				AccountNumber:  2,
				AdditionalInfo: `{"description":"Charlie Brown","user_name":"101"}`,
			},
		},
	}

	phones := []models.Phone{ph1, ph2}

	// 3. Extract and sort directory contacts
	byNumber, byName := m.getSortedDirectory(phones)

	// We expect 3 contacts:
	// - Alice Smith (102)
	// - Bob Jones (103)
	// - Charlie Brown (101)
	if len(byNumber) != 3 {
		t.Fatalf("Expected 3 contacts, got %d", len(byNumber))
	}
	if len(byName) != 3 {
		t.Fatalf("Expected 3 contacts, got %d", len(byName))
	}

	// 4. Verify sorting by phone number
	// Expected: Charlie (101), Alice (102), Bob (103)
	if byNumber[0].PhoneNumber != "101" || byNumber[0].FirstName != "Charlie" {
		t.Errorf("Expected first entry sorted by number to be Charlie (101), got %s (%s)", byNumber[0].FirstName, byNumber[0].PhoneNumber)
	}
	if byNumber[1].PhoneNumber != "102" || byNumber[1].FirstName != "Alice" {
		t.Errorf("Expected second entry sorted by number to be Alice (102), got %s (%s)", byNumber[1].FirstName, byNumber[1].PhoneNumber)
	}
	if byNumber[2].PhoneNumber != "103" || byNumber[2].FirstName != "Bob" {
		t.Errorf("Expected third entry sorted by number to be Bob (103), got %s (%s)", byNumber[2].FirstName, byNumber[2].PhoneNumber)
	}

	// 5. Verify sorting by name
	// Expected order: Alice, Bob, Charlie
	if byName[0].FirstName != "Alice" {
		t.Errorf("Expected first entry sorted by name to be Alice, got %s", byName[0].FirstName)
	}
	if byName[1].FirstName != "Bob" {
		t.Errorf("Expected second entry sorted by name to be Bob, got %s", byName[1].FirstName)
	}
	if byName[2].FirstName != "Charlie" {
		t.Errorf("Expected third entry sorted by name to be Charlie, got %s", byName[2].FirstName)
	}
}

func TestRenderTemplates(t *testing.T) {
	// Setup domain
	d := config.DomainSettings{
		Name: "test.local",
		Variables: map[string]string{
			"company": "Acme Inc.",
			"email":   "info@acme.com",
		},
	}

	// Prepare phones
	mac := "001122334455"
	phones := []models.Phone{
		{
			Domain:      "test.local",
			Type:        "phone",
			MacAddress:  &mac,
			PhoneNumber: func() *string { s := "102"; return &s }(),
			Description: "Alice Smith",
		},
	}

	m := NewManager(&config.SystemConfig{
		Domains: []config.DomainSettings{d},
	})

	byNumber, byName := m.getSortedDirectory(phones)

	allDomains := []map[string]interface{}{
		{
			"name":                d.Name,
			"variables":           d.Variables,
			"phones":              phones,
			"directory_by_number": byNumber,
			"directory_by_name":   byName,
		},
	}

	ctx := pongo2.Context{
		"all_domains": allDomains,
	}

	// Load templates from disk
	templateDir := filepath.Join("..", "..", "..", "conf", "vendors", "mitel", "directory")
	
	tplNumber, err := pongo2.FromFile(filepath.Join(templateDir, "directory_by_number.csv.tpl"))
	if err != nil {
		t.Fatalf("Failed to load directory_by_number.csv.tpl: %v", err)
	}

	outNumber, err := tplNumber.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute directory_by_number template: %v", err)
	}

	expectedNumber := "Alice,Smith,Acme Inc.,,,,,,,,,,,,info@acme.com,,,1,1,1,102,,,,\n"
	if outNumber != expectedNumber {
		t.Errorf("Unexpected output from directory_by_number.\nExpected:\n%q\nGot:\n%q", expectedNumber, outNumber)
	}

	tplName, err := pongo2.FromFile(filepath.Join(templateDir, "directory_by_name.csv.tpl"))
	if err != nil {
		t.Fatalf("Failed to load directory_by_name.csv.tpl: %v", err)
	}

	outName, err := tplName.Execute(ctx)
	if err != nil {
		t.Fatalf("Failed to execute directory_by_name template: %v", err)
	}

	if outName != expectedNumber {
		t.Errorf("Unexpected output from directory_by_name.\nExpected:\n%q\nGot:\n%q", expectedNumber, outName)
	}
}

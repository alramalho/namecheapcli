package namecheap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantSLD string
		wantTLD string
		wantErr bool
	}{
		{
			name:    "second level domain",
			domain:  "example.com",
			wantSLD: "example",
			wantTLD: "com",
		},
		{
			name:    "multi label sld",
			domain:  "example.co.uk",
			wantSLD: "example.co",
			wantTLD: "uk",
		},
		{
			name:    "invalid",
			domain:  "example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSLD, gotTLD, err := SplitDomain(tt.domain)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if gotSLD != tt.wantSLD || gotTLD != tt.wantTLD {
				t.Fatalf("SplitDomain(%q) = %q, %q; want %q, %q", tt.domain, gotSLD, gotTLD, tt.wantSLD, tt.wantTLD)
			}
		})
	}
}

func TestConfigFromEnvReadsHomeConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("NAMECHEAP_API_USER", "")
	t.Setenv("NAMECHEAP_API_KEY", "")
	t.Setenv("NAMECHEAP_USERNAME", "")
	t.Setenv("NAMECHEAP_CLIENT_IP", "")
	t.Setenv("NAMECHEAP_ENDPOINT", "")

	configPath := filepath.Join(home, ".namecheapcli")
	if err := os.WriteFile(configPath, []byte(`
NAMECHEAP_API_USER=file-user
NAMECHEAP_API_KEY=file-key
NAMECHEAP_USERNAME=file-username
NAMECHEAP_CLIENT_IP=203.0.113.10
`), 0600); err != nil {
		t.Fatal(err)
	}

	config := ConfigFromEnv(true)

	if config.APIUser != "file-user" {
		t.Fatalf("APIUser = %q; want file-user", config.APIUser)
	}
	if config.APIKey != "file-key" {
		t.Fatalf("APIKey = %q; want file-key", config.APIKey)
	}
	if config.UserName != "file-username" {
		t.Fatalf("UserName = %q; want file-username", config.UserName)
	}
	if config.ClientIP != "203.0.113.10" {
		t.Fatalf("ClientIP = %q; want 203.0.113.10", config.ClientIP)
	}
	if config.Endpoint != SandboxEndpoint {
		t.Fatalf("Endpoint = %q; want sandbox endpoint", config.Endpoint)
	}
}

func TestConfigFromEnvOverridesHomeConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("NAMECHEAP_API_USER", "env-user")
	t.Setenv("NAMECHEAP_API_KEY", "")
	t.Setenv("NAMECHEAP_USERNAME", "")
	t.Setenv("NAMECHEAP_CLIENT_IP", "")
	t.Setenv("NAMECHEAP_ENDPOINT", "https://example.test/xml.response")

	configPath := filepath.Join(home, ".namecheapcli")
	if err := os.WriteFile(configPath, []byte(`
API_USER=file-user
API_KEY=file-key
USERNAME=file-username
CLIENT_IP=203.0.113.10
`), 0600); err != nil {
		t.Fatal(err)
	}

	config := ConfigFromEnv(true)

	if config.APIUser != "env-user" {
		t.Fatalf("APIUser = %q; want env-user", config.APIUser)
	}
	if config.APIKey != "file-key" {
		t.Fatalf("APIKey = %q; want file-key", config.APIKey)
	}
	if config.UserName != "file-username" {
		t.Fatalf("UserName = %q; want file-username", config.UserName)
	}
	if config.ClientIP != "203.0.113.10" {
		t.Fatalf("ClientIP = %q; want 203.0.113.10", config.ClientIP)
	}
	if config.Endpoint != "https://example.test/xml.response" {
		t.Fatalf("Endpoint = %q; want env endpoint", config.Endpoint)
	}
}

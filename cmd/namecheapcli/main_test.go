package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alramalho/namecheapcli/internal/namecheap"
)

func TestSetHostUpdatesMatchingNameAndType(t *testing.T) {
	hosts := []namecheap.Host{
		{Name: "@", Type: "A", Address: "192.0.2.1", TTL: "1800"},
		{Name: "www", Type: "CNAME", Address: "old.example.com", TTL: "1800"},
	}

	got, err := setHost(hosts, namecheap.Host{Name: "www", Type: "CNAME", Address: "new.example.com", TTL: "300"})
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d; want 2", len(got))
	}
	if got[1].Address != "new.example.com" || got[1].TTL != "300" {
		t.Fatalf("record was not updated: %+v", got[1])
	}
}

func TestSetHostAppendsWhenNoMatch(t *testing.T) {
	hosts := []namecheap.Host{{Name: "@", Type: "A", Address: "192.0.2.1", TTL: "1800"}}

	got, err := setHost(hosts, namecheap.Host{Name: "www", Type: "CNAME", Address: "example.com", TTL: "300"})
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d; want 2", len(got))
	}
}

func TestParseMutationOptionsAcceptsFlagsAfterPositionals(t *testing.T) {
	pos, opts, err := parseMutationOptions([]string{
		"example.com", "A", "@", "203.0.113.10", "--ttl", "300", "--mx-pref=10", "--dry-run",
	})
	if err != nil {
		t.Fatal(err)
	}

	if opts.ttl != "300" || opts.mxPref != "10" || !opts.dryRun {
		t.Fatalf("opts = %+v", opts)
	}
	want := []string{"example.com", "A", "@", "203.0.113.10"}
	for i := range want {
		if pos[i] != want[i] {
			t.Fatalf("pos[%d] = %q; want %q", i, pos[i], want[i])
		}
	}
}

func TestRunConfigureWritesHomeConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out bytes.Buffer
	err := runConfigure([]string{
		"--api-user", "configured-user",
		"--api-key", "configured-key",
		"--username", "configured-username",
		"--client-ip", "203.0.113.20",
	}, strings.NewReader(""), &out)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(home, ".namecheapcli")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("mode = %v; want 0600", got)
	}
	if !strings.Contains(out.String(), path) {
		t.Fatalf("output = %q; want config path", out.String())
	}

	config := namecheap.ConfigFromEnv(false)
	if config.APIUser != "configured-user" {
		t.Fatalf("APIUser = %q; want configured-user", config.APIUser)
	}
	if config.APIKey != "configured-key" {
		t.Fatalf("APIKey = %q; want configured-key", config.APIKey)
	}
	if config.UserName != "configured-username" {
		t.Fatalf("UserName = %q; want configured-username", config.UserName)
	}
	if config.ClientIP != "203.0.113.20" {
		t.Fatalf("ClientIP = %q; want 203.0.113.20", config.ClientIP)
	}
}

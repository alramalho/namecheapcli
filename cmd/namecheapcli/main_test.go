package main

import (
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

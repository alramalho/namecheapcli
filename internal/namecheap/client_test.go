package namecheap

import "testing"

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

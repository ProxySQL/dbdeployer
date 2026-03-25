package postgresql

import "testing"

func TestParseDebVersion(t *testing.T) {
	tests := []struct {
		filename string
		wantVer  string
		wantErr  bool
	}{
		{"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb", "16.13", false},
		{"postgresql-17_17.2-1_amd64.deb", "17.2", false},
		{"postgresql-client-16_16.13-0ubuntu0.24.04.1_amd64.deb", "16.13", false},
		{"random-file.tar.gz", "", true},
		{"postgresql-16_bad-version.deb", "", true},
	}
	for _, tt := range tests {
		ver, err := ParseDebVersion(tt.filename)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDebVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			continue
		}
		if ver != tt.wantVer {
			t.Errorf("ParseDebVersion(%q) = %q, want %q", tt.filename, ver, tt.wantVer)
		}
	}
}

func TestClassifyDebs(t *testing.T) {
	files := []string{
		"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb",
		"postgresql-client-16_16.13-0ubuntu0.24.04.1_amd64.deb",
	}
	server, client, err := ClassifyDebs(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server != files[0] {
		t.Errorf("server = %q, want %q", server, files[0])
	}
	if client != files[1] {
		t.Errorf("client = %q, want %q", client, files[1])
	}
}

func TestClassifyDebsMissingClient(t *testing.T) {
	files := []string{"postgresql-16_16.13-0ubuntu0.24.04.1_amd64.deb"}
	_, _, err := ClassifyDebs(files)
	if err == nil {
		t.Error("expected error for missing client deb")
	}
}

func TestRequiredBinaries(t *testing.T) {
	expected := []string{"postgres", "initdb", "pg_ctl", "psql", "pg_basebackup"}
	got := RequiredBinaries()
	if len(got) != len(expected) {
		t.Fatalf("expected %d binaries, got %d", len(expected), len(got))
	}
	for i, name := range expected {
		if got[i] != name {
			t.Errorf("binary[%d] = %q, want %q", i, got[i], name)
		}
	}
}

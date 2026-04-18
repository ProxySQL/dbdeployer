package postgresql

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeGithubRelease mirrors the February 2026 Postgres.app release (v2.9.4)
// which has five single-version .dmg files plus one multi-version bundle.
// Using a fake server keeps the test offline and deterministic.
var fakeGithubRelease = githubRelease{
	TagName: "v2.9.4",
	Name:    "February 2026 Bugfix Updates (Feb 26)",
	Assets: []githubAsset{
		{
			Name:               "Postgres-2.9.4-14-15-16-17-18.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-14-15-16-17-18.dmg",
			Size:               532 * 1024 * 1024,
		},
		{
			Name:               "Postgres-2.9.4-14.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-14.dmg",
			Size:               84 * 1024 * 1024,
		},
		{
			Name:               "Postgres-2.9.4-15.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-15.dmg",
			Size:               99 * 1024 * 1024,
		},
		{
			Name:               "Postgres-2.9.4-16.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-16.dmg",
			Size:               107 * 1024 * 1024,
		},
		{
			Name:               "Postgres-2.9.4-17.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-17.dmg",
			Size:               113 * 1024 * 1024,
		},
		{
			Name:               "Postgres-2.9.4-18.dmg",
			BrowserDownloadURL: "https://example.invalid/Postgres-2.9.4-18.dmg",
			Size:               116 * 1024 * 1024,
		},
	},
}

// TestLatestPostgresAppAssetsParsing verifies that the asset filter picks
// up only single-version .dmg files, drops the multi-version bundle, and
// returns them sorted newest-major-first.
func TestLatestPostgresAppAssetsParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeGithubRelease)
	}))
	defer server.Close()

	assets, err := parseAssetsFromURL(server.URL)
	if err != nil {
		t.Fatalf("parseAssetsFromURL: %v", err)
	}

	wantMajors := []string{"18", "17", "16", "15", "14"}
	if len(assets) != len(wantMajors) {
		t.Fatalf("expected %d single-version assets, got %d: %+v",
			len(wantMajors), len(assets), assets)
	}
	for i, want := range wantMajors {
		if assets[i].Major != want {
			t.Errorf("assets[%d].Major = %q, want %q", i, assets[i].Major, want)
		}
		if assets[i].AppVersion != "2.9.4" {
			t.Errorf("assets[%d].AppVersion = %q, want %q", i, assets[i].AppVersion, "2.9.4")
		}
	}

	// The multi-version bundle must not appear.
	for _, a := range assets {
		if a.Major == "" || len(a.Major) > 2 {
			t.Errorf("unexpected asset slipped through filter: %+v", a)
		}
	}
}

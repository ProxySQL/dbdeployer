package downloads

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildMariaDBAPIURL(t *testing.T) {
	oldBase := MariaDBRestAPIBaseURL
	MariaDBRestAPIBaseURL = "https://downloads.mariadb.org/rest-api/mariadb"
	defer func() { MariaDBRestAPIBaseURL = oldBase }()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "short version", version: "11.4", want: "https://downloads.mariadb.org/rest-api/mariadb/11.4/latest/"},
		{name: "full version", version: "11.4.10", want: "https://downloads.mariadb.org/rest-api/mariadb/11.4.10/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMariaDBAPIURL(tt.version)
			if got != tt.want {
				t.Fatalf("buildMariaDBAPIURL() got %q want %q", got, tt.want)
			}
		})
	}
}

func TestGetMariaDBTarballFromAPI(t *testing.T) {
	payload := `{
		"releases": {
			"11.4.10": {
				"release_id": "11.4.10",
				"files": [
					{
						"file_name": "mariadb-11.4.10-linux-systemd-x86_64.tar.gz",
						"package_type": "gzipped tar file",
						"os": "Linux",
						"cpu": "x86_64",
						"checksum": {
							"md5sum": "",
							"sha1sum": "",
							"sha256sum": "",
							"sha512sum": "abc123"
						},
						"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.4.10/mariadb-11.4.10-linux-systemd-x86_64.tar.gz"
					},
					{
						"file_name": "mariadb-11.4.10.tar.gz",
						"package_type": "gzipped tar file",
						"os": "Source",
						"cpu": null,
						"checksum": {
							"md5sum": "",
							"sha1sum": "",
							"sha256sum": "",
							"sha512sum": "def456"
						},
						"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.4.10/mariadb-11.4.10.tar.gz"
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer server.Close()

	oldBase := MariaDBRestAPIBaseURL
	MariaDBRestAPIBaseURL = server.URL + "/mariadb"
	defer func() { MariaDBRestAPIBaseURL = oldBase }()

	tbd, err := GetMariaDBTarballFromAPI("11.4", "linux", "amd64", false)
	if err != nil {
		t.Fatalf("GetMariaDBTarballFromAPI() unexpected error: %v", err)
	}

	if tbd.Name != "mariadb-11.4.10-linux-systemd-x86_64.tar.gz" {
		t.Fatalf("unexpected Name: %q", tbd.Name)
	}
	if tbd.Url != "https://downloads.mariadb.org/rest-api/mariadb/11.4.10/mariadb-11.4.10-linux-systemd-x86_64.tar.gz" {
		t.Fatalf("unexpected Url: %q", tbd.Url)
	}
	if tbd.Checksum != "SHA512:abc123" {
		t.Fatalf("unexpected Checksum: %q", tbd.Checksum)
	}
	if tbd.OperatingSystem != "Linux" {
		t.Fatalf("unexpected OperatingSystem: %q", tbd.OperatingSystem)
	}
	if tbd.Arch != "amd64" {
		t.Fatalf("unexpected Arch: %q", tbd.Arch)
	}
	if tbd.Flavor != "mariadb" {
		t.Fatalf("unexpected Flavor: %q", tbd.Flavor)
	}
	if tbd.ShortVersion != "11.4" {
		t.Fatalf("unexpected ShortVersion: %q", tbd.ShortVersion)
	}
	if tbd.Version != "11.4.10" {
		t.Fatalf("unexpected Version: %q", tbd.Version)
	}
}

func TestGetMariaDBTarballFromAPI_DarwinFallbackToSource(t *testing.T) {
	payload := `{
		"releases": {
			"11.4.10": {
				"release_id": "11.4.10",
				"files": [
					{
						"file_name": "mariadb-11.4.10.tar.gz",
						"package_type": "gzipped tar file",
						"os": "Source",
						"cpu": null,
						"checksum": {
							"md5sum": "",
							"sha1sum": "",
							"sha256sum": "",
							"sha512sum": "def456"
						},
						"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.4.10/mariadb-11.4.10.tar.gz"
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer server.Close()

	oldBase := MariaDBRestAPIBaseURL
	MariaDBRestAPIBaseURL = server.URL + "/mariadb"
	defer func() { MariaDBRestAPIBaseURL = oldBase }()

	tbd, err := GetMariaDBTarballFromAPI("11.4", "darwin", "amd64", false)
	if err != nil {
		t.Fatalf("GetMariaDBTarballFromAPI() unexpected error: %v", err)
	}

	if tbd.Name != "mariadb-11.4.10.tar.gz" {
		t.Fatalf("unexpected Name: %q", tbd.Name)
	}
}

func TestGetMariaDBTarballFromAPI_ReleaseDataKey(t *testing.T) {
	payload := `{
		"release_data": {
			"11.7.2": {
				"release_id": "11.7.2",
				"files": [
					{
						"file_name": "mariadb-11.7.2-linux-systemd-x86_64.tar.gz",
						"package_type": "gzipped tar file",
						"os": "Linux",
						"cpu": "x86_64",
						"checksum": {
							"md5sum": null,
							"sha1sum": null,
							"sha256sum": null,
							"sha512sum": "a527"
						},
						"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.7.2/mariadb-11.7.2-linux-systemd-x86_64.tar.gz"
					},
					{
						"file_name": "yum/",
						"package_type": null,
						"os": null,
						"cpu": null,
						"checksum": {
							"md5sum": null,
							"sha1sum": null,
							"sha256sum": null,
							"sha512sum": null
						},
						"file_download_url": "http://downloads.mariadb.org/rest-api/mariadb/11.7.2/yum/"
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer server.Close()

	oldBase := MariaDBRestAPIBaseURL
	MariaDBRestAPIBaseURL = server.URL + "/mariadb"
	defer func() { MariaDBRestAPIBaseURL = oldBase }()

	tbd, err := GetMariaDBTarballFromAPI("11.7.2", "linux", "amd64", false)
	if err != nil {
		t.Fatalf("GetMariaDBTarballFromAPI() unexpected error: %v", err)
	}

	if tbd.Name != "mariadb-11.7.2-linux-systemd-x86_64.tar.gz" {
		t.Fatalf("unexpected Name: %q", tbd.Name)
	}
	if tbd.Checksum != "SHA512:a527" {
		t.Fatalf("unexpected Checksum: %q", tbd.Checksum)
	}
}

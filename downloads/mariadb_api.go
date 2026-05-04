// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
// Copyright © 2026 Frédéric Descamps - lefred
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package downloads

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
)

const mariaDBProductID = "mariadb"

// MariaDBRestAPIBaseURL is exported for tests and defaults to the official MariaDB downloads API.
var MariaDBRestAPIBaseURL = "https://downloads.mariadb.org/rest-api/" + mariaDBProductID

var mariaDBFullVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type mariaDBAPIChecksum struct {
	MD5Sum    *string `json:"md5sum"`
	SHA1Sum   *string `json:"sha1sum"`
	SHA256Sum *string `json:"sha256sum"`
	SHA512Sum *string `json:"sha512sum"`
}

type mariaDBAPIFile struct {
	FileName        string             `json:"file_name"`
	PackageType     string             `json:"package_type"`
	OS              *string            `json:"os"`
	CPU             *string            `json:"cpu"`
	Checksum        mariaDBAPIChecksum `json:"checksum"`
	FileDownloadURL string             `json:"file_download_url"`
}

type mariaDBAPIRelease struct {
	ReleaseID string           `json:"release_id"`
	Files     []mariaDBAPIFile `json:"files"`
}

type mariaDBAPIResponse struct {
	Releases    map[string]mariaDBAPIRelease `json:"releases"`
	ReleaseData map[string]mariaDBAPIRelease `json:"release_data"`
}

func (r mariaDBAPIResponse) releaseMap() map[string]mariaDBAPIRelease {
	if len(r.Releases) > 0 {
		return r.Releases
	}
	return r.ReleaseData
}

// GetMariaDBTarballFromAPI returns a new TarballDescription for a MariaDB release
// discovered via the MariaDB downloads REST API.
func GetMariaDBTarballFromAPI(version, OS, arch string, minimal bool) (TarballDescription, error) {
	url := buildMariaDBAPIURL(version)
	apiResponse, err := getMariaDBAPIResponse(url)
	if err != nil {
		return TarballDescription{}, err
	}

	releaseID, releaseInfo, err := getBestReleaseFromMariaDBResponse(apiResponse)
	if err != nil {
		return TarballDescription{}, err
	}

	apiFile, err := findBestMariaDBFile(releaseInfo.Files, OS, arch, minimal)
	if err != nil {
		return TarballDescription{}, fmt.Errorf("no matching MariaDB file for release %s: %s", releaseID, err)
	}

	tbd, err := getTarballDescriptionFromMariaDBFile(apiFile)
	if err != nil {
		return TarballDescription{}, err
	}

	if tbd.Version == "" {
		tbd.Version = releaseID
	}
	if tbd.ShortVersion == "" {
		releaseVersionList, shortErr := common.VersionToList(releaseID)
		if shortErr == nil && len(releaseVersionList) >= 2 {
			tbd.ShortVersion = fmt.Sprintf("%d.%d", releaseVersionList[0], releaseVersionList[1])
		}
	}

	return tbd, nil
}

func buildMariaDBAPIURL(version string) string {
	base := strings.TrimSuffix(MariaDBRestAPIBaseURL, "/")
	if mariaDBFullVersionRE.MatchString(version) {
		return fmt.Sprintf("%s/%s/", base, version)
	}
	return fmt.Sprintf("%s/%s/latest/", base, version)
}

func getMariaDBAPIResponse(url string) (mariaDBAPIResponse, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	// #nosec G107
	resp, err := client.Get(url)
	if err != nil {
		return mariaDBAPIResponse{}, fmt.Errorf("error querying MariaDB API at %s: %s", url, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return mariaDBAPIResponse{}, fmt.Errorf("MariaDB API returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mariaDBAPIResponse{}, fmt.Errorf("error reading MariaDB API response: %s", err)
	}

	var apiResponse mariaDBAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return mariaDBAPIResponse{}, fmt.Errorf("error decoding MariaDB API response: %s", err)
	}

	if len(apiResponse.releaseMap()) == 0 {
		return mariaDBAPIResponse{}, fmt.Errorf("MariaDB API response did not contain releases or release_data")
	}

	return apiResponse, nil
}

func getBestReleaseFromMariaDBResponse(apiResponse mariaDBAPIResponse) (string, mariaDBAPIRelease, error) {
	bestReleaseID := ""
	bestRelease := mariaDBAPIRelease{}
	bestVersionList := []int{0, 0, 0}
	releaseMap := apiResponse.releaseMap()

	for releaseID, releaseInfo := range releaseMap {
		releaseVersionList, err := common.VersionToList(releaseID)
		if err != nil {
			continue
		}
		greaterOrEqual, err := common.GreaterOrEqualVersionList(releaseVersionList, bestVersionList)
		if err != nil {
			continue
		}
		if greaterOrEqual {
			bestReleaseID = releaseID
			bestRelease = releaseInfo
			bestVersionList = releaseVersionList
		}
	}

	if bestReleaseID == "" {
		return "", mariaDBAPIRelease{}, fmt.Errorf("could not determine release from MariaDB API response")
	}

	return bestReleaseID, bestRelease, nil
}

func findBestMariaDBFile(files []mariaDBAPIFile, OS, arch string, minimal bool) (mariaDBAPIFile, error) {
	if len(files) == 0 {
		return mariaDBAPIFile{}, fmt.Errorf("no files in release")
	}

	requestedOS := normalizeMariaDBOS(OS)
	requestedArch := normalizeMariaDBArch(arch)

	var matches []mariaDBAPIFile
	for _, fileInfo := range files {
		if !isMariaDBTarball(fileInfo.FileName) {
			continue
		}
		if minimal && !strings.Contains(strings.ToLower(fileInfo.FileName), "minimal") {
			continue
		}
		fileOS := valueOrEmpty(fileInfo.OS)
		if requestedOS != "" && !strings.EqualFold(fileOS, requestedOS) {
			continue
		}
		fileCPU := normalizeMariaDBArch(valueOrEmpty(fileInfo.CPU))
		if requestedArch != "" && fileCPU != "" && fileCPU != requestedArch {
			continue
		}
		matches = append(matches, fileInfo)
	}

	if len(matches) == 0 && requestedOS == "darwin" {
		for _, fileInfo := range files {
			if isMariaDBTarball(fileInfo.FileName) && strings.EqualFold(valueOrEmpty(fileInfo.OS), "Source") {
				matches = append(matches, fileInfo)
			}
		}
	}

	if len(matches) == 0 {
		return mariaDBAPIFile{}, fmt.Errorf("no tarball matches OS=%s arch=%s minimal=%v", OS, arch, minimal)
	}

	for _, fileInfo := range matches {
		if strings.Contains(strings.ToLower(fileInfo.FileName), "systemd") {
			return fileInfo, nil
		}
	}

	return matches[0], nil
}

func getTarballDescriptionFromMariaDBFile(apiFile mariaDBAPIFile) (TarballDescription, error) {
	if apiFile.FileName == "" {
		return TarballDescription{}, fmt.Errorf("empty filename in MariaDB API response")
	}
	if apiFile.FileDownloadURL == "" {
		return TarballDescription{}, fmt.Errorf("empty download URL in MariaDB API response")
	}

	flavor, version, shortVersion, err := common.FindTarballInfo(apiFile.FileName)
	if err != nil {
		flavor = "mariadb"
	}

	return TarballDescription{
		Name:            apiFile.FileName,
		Checksum:        getPreferredMariaDBChecksum(apiFile.Checksum),
		OperatingSystem: normalizeDBDeployerOS(valueOrEmpty(apiFile.OS)),
		Arch:            normalizeMariaDBArch(valueOrEmpty(apiFile.CPU)),
		Url:             normalizeMariaDBDownloadURL(apiFile.FileDownloadURL),
		Flavor:          flavor,
		Minimal:         strings.Contains(strings.ToLower(apiFile.FileName), "minimal"),
		ShortVersion:    shortVersion,
		Version:         version,
	}, nil
}

func getPreferredMariaDBChecksum(checksum mariaDBAPIChecksum) string {
	if checksum.SHA512Sum != nil && *checksum.SHA512Sum != "" {
		return "SHA512:" + *checksum.SHA512Sum
	}
	if checksum.SHA256Sum != nil && *checksum.SHA256Sum != "" {
		return "SHA256:" + *checksum.SHA256Sum
	}
	if checksum.SHA1Sum != nil && *checksum.SHA1Sum != "" {
		return "SHA1:" + *checksum.SHA1Sum
	}
	if checksum.MD5Sum != nil && *checksum.MD5Sum != "" {
		return "MD5:" + *checksum.MD5Sum
	}
	return ""
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func normalizeMariaDBDownloadURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		return "https://" + strings.TrimPrefix(url, "http://")
	}
	return url
}

func normalizeDBDeployerOS(osName string) string {
	switch strings.ToLower(osName) {
	case "linux":
		return "Linux"
	case "darwin", "macos", "osx":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return osName
	}
}

func normalizeMariaDBOS(OS string) string {
	switch strings.ToLower(OS) {
	case "linux":
		return "Linux"
	case "darwin", "macos", "osx":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return OS
	}
}

func normalizeMariaDBArch(arch string) string {
	switch strings.ToLower(arch) {
	case "x86_64", "x86-64", "amd64":
		return "amd64"
	case "aarch64", "arm64":
		return "arm64"
	default:
		return arch
	}
}

func isMariaDBTarball(fileName string) bool {
	lower := strings.ToLower(fileName)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".tgz") {
		return true
	}
	return false
}

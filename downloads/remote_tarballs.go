// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ProxySQL/dbdeployer/common"
	"jaytaylor.com/html2text"
)

type TarballDef struct {
	Flavor      string
	NameInFile  string
	NameInUrl   string
	DownloadDir string
}
type TarballType string

const (
	// Recognized Operating systems

	OsLinux = "Linux"
	OsMacOs = "MacOs"

	preferredUserAgentLinux           = "dbdeployer/XXXX (X11; Linux; rv:XXXX) dbdeployer/XXXX"
	preferredUserAgentMacOs           = "dbdeployer/XXXX (Macintosh; Mac OS; rv:XXXX) dbdeployer/XXXX"
	alternativeUserAgentLinuxTemplate = "Mozilla/5.0 (X11; Linux ; rv:XXXX.0) Gecko/20100101 Firefox/XXXX.0"
	alternativeUserAgentMacOsTemplate = "Mozilla/5.0 (Macintosh; Mac OS; rv:XXXX.0) Gecko/20100101 Firefox/XXXX.0"
	basePageUrl                       = "https://dev.mysql.com/downloads"
	baseDownloadUrl                   = "https://dev.mysql.com/get/Downloads"

	// Tarball types

	TtMysql   TarballType = "mysql"
	TtCluster TarballType = "cluster"
	TtShell   TarballType = "shell"
)

var downloadsSettings = map[TarballType]TarballDef{
	TtCluster: {
		Flavor:      "ndb",
		NameInFile:  "mysql-cluster",
		NameInUrl:   "cluster",
		DownloadDir: "MySQL-Cluster-VERSION",
	},
	TtMysql: {
		Flavor:      "mysql",
		NameInFile:  "mysql",
		NameInUrl:   "mysql",
		DownloadDir: "MySQL-VERSION",
	},
	TtShell: {
		Flavor:      "shell",
		NameInFile:  "mysql-shell",
		NameInUrl:   "shell",
		DownloadDir: "MySQL-Shell-VERSION",
	},
}

var reReleaseSeries = regexp.MustCompile(`^\d+\.\d+$`)

var (
	userAgentLinux = strings.Replace(preferredUserAgentLinux, "XXXX", common.VersionDef, -1)
	userAgentMacOs = strings.Replace(preferredUserAgentMacOs, "XXXX", common.VersionDef, -1)

	thisYear = fmt.Sprintf("%d", time.Now().Year())
	// Alternative user agents
	alternativeUserAgentMacOs = strings.Replace(alternativeUserAgentMacOsTemplate, "XXXX", thisYear, -1)
	alternativeUserAgentLinux = strings.Replace(alternativeUserAgentLinuxTemplate, "XXXX", thisYear, -1)
)

var userAgents = map[string]string{
	OsLinux: userAgentLinux,
	OsMacOs: userAgentMacOs,
}

var internalOsName = map[string]string{
	OsLinux: "Linux",
	OsMacOs: "Darwin",
}
var osNormalize = map[string]string{
	OsLinux:  OsLinux,
	"linux":  OsLinux,
	OsMacOs:  OsMacOs,
	"macos":  OsMacOs,
	"macOs":  OsMacOs,
	"macOS":  OsMacOs,
	"macOSX": OsMacOs,
	"Macos":  OsMacOs,
	"osx":    OsMacOs,
	"OSX":    OsMacOs,
	"osX":    OsMacOs,
	"darwin": OsMacOs,
	"Darwin": OsMacOs,
}

func validateTarballRequest(tarballType TarballType, version string, OS string) (string, error) {
	osText, ok := osNormalize[OS]
	if !ok {
		return "", fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}
	OS = osText
	if _, foundOs := userAgents[osText]; !foundOs {
		return "", fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}

	settings, foundSettings := downloadsSettings[tarballType]
	if !foundSettings {
		return "", fmt.Errorf("unrecognized tarball type %s: it must be one of [%s, %s, %s] ", OS, TtMysql, TtCluster, TtShell)
	}
	// Any release series is accepted, as long as it has the form "major.minor".
	// Whether the series is actually available is decided by the download
	// site, and verified in parseRemoteTarballPage.
	if !reReleaseSeries.MatchString(version) {
		return "", fmt.Errorf("version '%s' is not accepted for tarball type %s: it must be a release series in the form 'major.minor' (e.g. 8.4)", version, settings.Flavor)
	}
	return OS, nil
}

func GetRemoteTarballList(tarballType TarballType, version, OS string, withSize, alternativeUserAgent bool) ([]TarballDescription, error) {
	var err error
	OS, err = validateTarballRequest(tarballType, version, OS)
	if err != nil {
		return nil, err
	}

	pageUrl := basePageUrl + "/" + downloadsSettings[tarballType].NameInUrl + "/" + version + ".html"

	if alternativeUserAgent {
		userAgents[OsLinux] = alternativeUserAgentLinux
		userAgents[OsMacOs] = alternativeUserAgentMacOs
	}

	userAgent, foundOs := userAgents[OS]
	if !foundOs {
		return nil, fmt.Errorf("unrecognized OS %s: it must be one of [%s, %s] ", OS, OsMacOs, OsLinux)
	}
	if common.IsEnvSet("SBDEBUG") {
		fmt.Printf("user agent: %s\n", userAgent)
	}

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, pageUrl, nil)
	req.Header.Add("User-Agent", userAgent)

	response, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET %s: %s", pageUrl, err)
	}

	body, err := io.ReadAll(response.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("[GetRemoteTarballList] error closing response body: %s", err)
		}
	}(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	htmlText := string(body)

	text, err := html2text.FromString(htmlText, html2text.Options{PrettyTables: true})
	if err != nil {
		return nil, err
	}

	return parseRemoteTarballPage(tarballType, version, OS, text, userAgent, withSize)
}

// parseRemoteTarballPage extracts the tarballs of the requested release series
// from the text of a downloads page.
// The download site does not reject a release series it no longer lists (or
// never had): it serves the page of the current default release instead.
// Only tarballs whose version belongs to the requested series are returned, so
// that a request for an unavailable series fails instead of silently
// registering tarballs of a different series.
func parseRemoteTarballPage(tarballType TarballType, version, OS, text, userAgent string, withSize bool) ([]TarballDescription, error) {
	var result []TarballDescription
	reVersions := regexp.MustCompile(`((\d+\.\d+)\.\d+)`)
	/*
		reLine := regexp.MustCompile(fmt.Sprintf(
			`\(` +                                   // An open parenthesis
				`(` +                                // start capture
				downloads[tarballType].NameInFile +  // the tarball type (mysql | cluster | shell)
				`-\d+\.\d+\.\d+` +                   // A version
				`[^\)]+` +                           // more characters except a closed parenthesis
				`\S+64`                             // x86-64 or x86_64 or amd64
				`[^\)]+` +                           // more characters except a closed parenthesis
				`z)` +                               // a "z" (as in .gz or .xz) followed by a end capture
				`\)` +                               // a closed parenthesis
				`.*?MD5:` +                          // any character up to "MD5:"
				`\s*` +                              // optional spaces
				`(\w+)`))                            // Capture a string of alphanumeric characters (the checksum)
	*/
	reLine := regexp.MustCompile(fmt.Sprintf(`\((%s-\d+\.\d+\.\d+[^\)]+-(\S+64)[^\)]+z)\).*?MD5:\s*(\w+)`, downloadsSettings[tarballType].NameInFile))

	matches := reLine.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		if common.IsEnvSet("SBDEBUG") {
			fmt.Printf("With user agent '%s'\n", userAgent)
			fmt.Printf("Text retrieved from server: %s\n", basePageUrl)
			fmt.Println(text)
		} else {
			return nil, fmt.Errorf("no %s tarballs found for %s - set environment variable 'SBDBUG' to see failure details", tarballType, version)
		}
		return nil, fmt.Errorf("no %s tarballs found for %s", tarballType, version)
	}

	var otherSeries []string
	for _, m := range matches {
		versionList := reVersions.FindAllStringSubmatch(m[1], 1)
		if len(versionList) == 0 || len(versionList[0]) < 2 {
			return nil, fmt.Errorf("error extracting version from file name '%s'", m[1])
		}
		longVersion := versionList[0][1]
		shortVersion := versionList[0][2]
		if shortVersion != version {
			if !slices.Contains(otherSeries, shortVersion) {
				otherSeries = append(otherSeries, shortVersion)
			}
			continue
		}

		tbd := TarballDescription{
			Name:            m[1],
			Checksum:        "MD5:" + m[3],
			OperatingSystem: internalOsName[OS],
			Arch:            archNormalize(m[2]),
			ShortVersion:    shortVersion,
			Version:         longVersion,
			Flavor:          downloadsSettings[tarballType].Flavor,
			Minimal:         strings.Contains(m[1], "minimal"),
			Url: baseDownloadUrl + "/" +
				strings.Replace(downloadsSettings[tarballType].DownloadDir, "VERSION", version, 1) +
				"/" + m[1],
		}
		if withSize {
			tbd.Size, _ = checkRemoteUrl(tbd.Url)
		}
		result = append(result, tbd)
	}
	if len(result) == 0 {
		err := fmt.Errorf("no %s tarballs found for release series %s: the download site lists only %v", tarballType, version, otherSeries)
		// Retired series are no longer listed by the download site, but the
		// registry may already have their releases.
		known := registrySeriesVersions(downloadsSettings[tarballType].Flavor, version)
		if len(known) > 0 {
			err = fmt.Errorf("%s. The tarball registry already has %s %v: use 'dbdeployer downloads get-by-version %s'",
				err, tarballType, known, version)
		}
		return nil, err
	}
	return result, nil
}

// registrySeriesVersions returns the versions of a release series found in
// the tarball registry for the given flavor.
func registrySeriesVersions(flavor, series string) []string {
	var versions []string
	for _, tb := range DefaultTarballRegistry.Tarballs {
		if tb.Flavor == flavor && tb.ShortVersion == series && !slices.Contains(versions, tb.Version) {
			versions = append(versions, tb.Version)
		}
	}
	return versions
}

// archNormalize converts the architecture found in a tarball name into the
// form used by the registry, which is matched against runtime.GOARCH.
func archNormalize(s string) string {
	switch s {
	case "x86_64", "x86-64":
		return "amd64"
	case "aarch64", "arm-64":
		return "arm64"
	}
	return s
}

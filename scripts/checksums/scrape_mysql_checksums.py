#!/usr/bin/env python3
"""
Scrape MD5 checksums from MySQL download archive pages and update
tarball_list.json in-place.

Usage: scrape_checksums.py <path-to-tarball_list.json>
"""
import json
import re
import sys
import subprocess
from pathlib import Path

ARCHIVE_URL = "https://downloads.mysql.com/archives/community/"
CURRENT_URL_TMPL = "https://dev.mysql.com/downloads/mysql/{major}.html"

# OS id in the form
OS_LINUX = "2"    # Linux - Generic
OS_MACOS = "33"   # macOS

# rowspan pattern:
#   <td class="sub-text">(<filename>)</td>
#   <td ...>MD5: <code class="md5"><hash></code>
ROW_RE = re.compile(
    r'\((?P<file>[^)]+\.tar\.(?:gz|xz))\).*?md5">(?P<md5>[0-9a-f]{32})</code>',
    re.DOTALL,
)


def fetch(url: str, cookies: str | None = None) -> str:
    """Fetch a URL via curl and return the body."""
    cmd = ["curl", "-sL"]
    if cookies:
        cmd += ["-b", cookies]
    cmd.append(url)
    return subprocess.check_output(cmd, text=True)


def prime_cookies() -> str:
    """Prime cookies by fetching the archives page once."""
    jar = "/tmp/mysql-scrape-cookies.txt"
    subprocess.check_call(
        ["curl", "-sL", "-c", jar, ARCHIVE_URL, "-o", "/dev/null"],
        stderr=subprocess.DEVNULL,
    )
    return jar


def scrape_page(url: str, cookies: str | None = None) -> dict[str, str]:
    """Extract {filename: md5} from a download page."""
    html = fetch(url, cookies)
    result: dict[str, str] = {}
    for m in ROW_RE.finditer(html):
        result[m.group("file")] = m.group("md5")
    return result


def scrape_archived(version: str, os_id: str, cookies: str) -> dict[str, str]:
    url = f"{ARCHIVE_URL}?tpl=version&os={os_id}&version={version}"
    return scrape_page(url, cookies)


def scrape_current(major: str, os_id: str) -> dict[str, str]:
    url = f"{CURRENT_URL_TMPL.format(major=major)}?os={os_id}"
    return scrape_page(url)


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <tarball_list.json>", file=sys.stderr)
        sys.exit(1)

    json_path = Path(sys.argv[1])
    data = json.loads(json_path.read_text())
    tarballs = data["Tarballs"]

    # Find MySQL entries needing checksums, grouped by version
    by_version: dict[str, list[dict]] = {}
    for tb in tarballs:
        if tb.get("flavor") != "mysql":
            continue
        if tb.get("checksum"):
            continue
        by_version.setdefault(tb["version"], []).append(tb)

    if not by_version:
        print("No MySQL entries need updating")
        return

    cookies = prime_cookies()
    updated = 0
    missing = []

    for version, entries in sorted(by_version.items()):
        major = ".".join(version.split(".")[:2])  # e.g. 8.4
        # Scrape both Linux and macOS pages — start with archived URL,
        # fall back to the current-GA page if the archived page serves
        # a different version (happens for the latest release).
        for os_id in (OS_LINUX, OS_MACOS):
            files = scrape_archived(version, os_id, cookies)
            # If we didn't find entries for this version, try current-GA page
            has_version = any(version in f for f in files)
            if not has_version:
                files = scrape_current(major, os_id)
            for tb in entries:
                name = tb["name"]
                if name in files:
                    tb["checksum"] = f"MD5:{files[name]}"
                    updated += 1

    # Second pass to collect still-missing entries
    for version, entries in by_version.items():
        for tb in entries:
            if not tb.get("checksum"):
                missing.append((version, tb["name"]))

    json_path.write_text(json.dumps(data, indent=2) + "\n")
    print(f"Updated {updated} entries")
    if missing:
        print(f"\nStill missing checksums for {len(missing)} files:")
        for v, name in sorted(missing)[:30]:
            print(f"  {v}  {name}")
        if len(missing) > 30:
            print(f"  ... and {len(missing) - 30} more")


if __name__ == "__main__":
    main()

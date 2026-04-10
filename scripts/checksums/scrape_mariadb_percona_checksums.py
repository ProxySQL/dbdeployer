#!/usr/bin/env python3
"""
Populate SHA256 checksums for MariaDB and Percona Server entries in
tarball_list.json.

MariaDB: archive.mariadb.org publishes sha256sums.txt in each directory.
Percona: downloads.percona.com publishes .sha256sum sidecar per file.

Usage: scrape_mariadb_percona.py <path-to-tarball_list.json>
"""
import json
import re
import subprocess
import sys
from pathlib import Path
from urllib.parse import urlparse


def fetch(url: str) -> tuple[int, str]:
    """Return (http_status, body) for a URL."""
    try:
        out = subprocess.run(
            ["curl", "-sL", "-w", "\n%{http_code}", url],
            capture_output=True, text=True, check=True,
        ).stdout
    except subprocess.CalledProcessError as e:
        return (0, "")
    # Last line is the HTTP code
    body, _, code = out.rpartition("\n")
    try:
        return (int(code), body)
    except ValueError:
        return (0, out)


def fetch_mariadb_dir(url: str) -> dict[str, str]:
    """For a MariaDB bintar URL, fetch the directory's sha256sums.txt and
    return {filename: sha256}."""
    # url: https://archive.mariadb.org/mariadb-10.11.9/bintar-linux-systemd-x86_64/mariadb-10.11.9-linux-systemd-x86_64.tar.gz
    parent = url.rsplit("/", 1)[0] + "/sha256sums.txt"
    code, body = fetch(parent)
    if code != 200:
        return {}
    result: dict[str, str] = {}
    for line in body.splitlines():
        # lines look like:  <hash>  ./<file>
        m = re.match(r"([0-9a-f]{64})\s+\.?/?(.+)$", line.strip())
        if m:
            result[m.group(2)] = m.group(1)
    return result


def fetch_percona_sum(url: str) -> str | None:
    """Fetch .sha256sum sidecar for a single tarball."""
    sum_url = url + ".sha256sum"
    code, body = fetch(sum_url)
    if code != 200:
        return None
    m = re.match(r"([0-9a-f]{64})\s+", body.strip())
    return m.group(1) if m else None


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <tarball_list.json>", file=sys.stderr)
        sys.exit(1)

    json_path = Path(sys.argv[1])
    data = json.loads(json_path.read_text())
    tarballs = data["Tarballs"]

    mariadb_dir_cache: dict[str, dict[str, str]] = {}

    updated = 0
    missing = []

    for tb in tarballs:
        if tb.get("checksum"):
            continue
        flavor = tb.get("flavor")
        name = tb["name"]
        url = tb.get("url", "")

        if flavor == "mariadb":
            # Cache directory listings
            dir_url = url.rsplit("/", 1)[0]
            if dir_url not in mariadb_dir_cache:
                mariadb_dir_cache[dir_url] = fetch_mariadb_dir(url)
            sums = mariadb_dir_cache[dir_url]
            if name in sums:
                tb["checksum"] = f"SHA256:{sums[name]}"
                updated += 1
            else:
                missing.append((flavor, tb["version"], name))

        elif flavor == "percona":
            sha = fetch_percona_sum(url)
            if sha:
                tb["checksum"] = f"SHA256:{sha}"
                updated += 1
            else:
                missing.append((flavor, tb["version"], name))

    json_path.write_text(json.dumps(data, indent=2) + "\n")
    print(f"Updated {updated} entries")
    if missing:
        print(f"\nStill missing: {len(missing)}")
        for row in missing[:20]:
            print(f"  {row}")


if __name__ == "__main__":
    main()

# Tarball checksum scrapers

These scripts populate the `checksum` field in `downloads/tarball_list.json`
by scraping the official publisher pages.

## Usage

```bash
# MySQL (MD5, scraped from dev.mysql.com download pages)
python3 scripts/checksums/scrape_mysql_checksums.py downloads/tarball_list.json

# MariaDB + Percona Server (SHA256, from upstream sidecar files)
python3 scripts/checksums/scrape_mariadb_percona_checksums.py downloads/tarball_list.json
```

Both scripts are idempotent: they skip entries that already have a checksum.
After running, verify there are no missing checksums other than TiDB (whose
"master" rolling tarballs are legitimately unversioned):

```bash
jq -r '.Tarballs[] | select(.checksum == null or .checksum == "") | "\(.flavor) \(.version) \(.name)"' \
  downloads/tarball_list.json
```

## When to re-run

Re-run these scripts after adding new MySQL / MariaDB / Percona versions
to `tarball_list.json`. A CI lint (tracked in [#84](https://github.com/ProxySQL/dbdeployer/issues/84))
will fail if new non-TiDB entries are added without a checksum.

## Upstream sources

- **MySQL**: MD5 scraped from
  `https://downloads.mysql.com/archives/community/?tpl=version&os=<id>&version=<X.Y.Z>`
  (fallback to `https://dev.mysql.com/downloads/mysql/<X.Y>.html` for the
  current-GA release). Requires a primed cookie jar (`/tmp/mysql-scrape-cookies.txt`).
- **MariaDB**: SHA256 read from the `sha256sums.txt` file in each
  `archive.mariadb.org/mariadb-X.Y.Z/bintar-.../` directory.
- **Percona Server**: SHA256 read from the per-tarball `.sha256sum` sidecar
  file on `downloads.percona.com`.

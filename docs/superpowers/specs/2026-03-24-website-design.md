# dbdeployer Website Design

**Date:** 2026-03-24
**Author:** Rene (ProxySQL)
**Status:** Draft

## Context

dbdeployer has rich documentation (44 wiki pages, 54 API versions, ProxySQL guide, PostgreSQL provider docs) but no proper website. The current setup is a default Jekyll theme on GitHub Pages rendering the README. The project is evolving from a MySQL-only sandbox tool into a multi-database infrastructure tool under ProxySQL's maintainership, and needs a web presence that reflects this.

## Goals

- **Primary audience:** MySQL/PostgreSQL developers searching for local sandbox/testing tools (SEO-first)
- **Secondary goal:** Introduce ProxySQL integration as a natural next step
- **Tone:** Documentation-focused with a commercial/marketing polish — not a corporate site, but professional enough to build confidence

## Tech Stack

- **Framework:** Astro with Starlight integration (Astro's official docs theme)
- **Why Starlight:** sidebar navigation, Pagefind search, dark/light mode, content collections, i18n-ready — all out of the box. Custom pages (landing, providers, blog) use standard Astro layouts outside Starlight.
- **Node.js:** 20 LTS (Astro 4.x requires Node 18.17+)
- **Hosting:** GitHub Pages
- **Deployment:** GitHub Actions → builds Astro → pushes to `gh-pages` branch
- **Base path:** `astro.config.mjs` must set `base: '/dbdeployer'` and `site: 'https://proxysql.github.io'` since this is a project repo (not org root)

## Project Structure

```
website/
  astro.config.mjs
  package.json
  src/
    content/
      config.ts       # Content collection schemas (docs + blog)
      docs/           # Starlight docs (migrated wiki pages)
      blog/           # Blog posts as .md files
    pages/
      index.astro     # Landing page (custom, not Starlight)
      providers.astro # Providers comparison page
      404.astro       # Custom 404 page (links back to home/docs)
      blog/
        index.astro   # Blog index (reverse-chronological list)
        [...slug].astro # Individual blog post pages
    components/       # Reusable Astro components (Hero, FeatureGrid, etc.)
    layouts/          # Custom layouts for landing/blog
    styles/           # Global CSS
  public/
    favicon.svg       # Site favicon
    og-image.png      # Default Open Graph image for social sharing
    images/           # Screenshots, diagrams
  scripts/
    copy-wiki.sh      # Build step: copies docs/wiki/ into src/content/docs/
```

Source lives in `website/` at the repo root. The `gh-pages` branch contains only the built output.

## Site Sections

### Home (Landing Page)

Custom `index.astro` — not a Starlight page. Marketing-oriented.

**Structure (top to bottom):**

1. **Nav bar** — logo/name, links: Getting Started, Docs, Providers, Blog, GitHub
2. **Hero section:**
   - Tagline: *"Deploy MySQL & PostgreSQL sandboxes in seconds"*
   - Subtitle: *"Create single instances, replication topologies, and full testing stacks — locally, without root, without Docker"*
   - CTAs: "Get Started" → quickstart guide, "View on GitHub" → repo
3. **Quick install snippet** — one-liner in a code block with copy button
4. **Feature grid** — 3-4 cards:
   - "Any Topology" — single, replication, group replication, fan-in, all-masters
   - "Multiple Databases" — MySQL, PostgreSQL, Percona, MariaDB
   - "ProxySQL Integration" — deploy read/write split stacks in one command
   - "No Root, No Docker" — runs entirely in userspace
5. **Terminal demo** — animated or static code block showing a deploy + connect flow
6. **Providers section** — brief cards for MySQL, PostgreSQL, ProxySQL linking to Providers page
7. **"What's New" strip** — latest 1-2 blog posts
8. **Footer** — links, GitHub, license

### Getting Started

Four polished, tutorial-style guides — **new content**, written fresh:

1. **Quick Start: MySQL Single** — install, deploy, connect, destroy
2. **Quick Start: MySQL Replication** — deploy replication, check status, test failover
3. **Quick Start: PostgreSQL** — unpack debs, deploy, connect via psql
4. **Quick Start: ProxySQL Integration** — deploy replication with `--with-proxysql`, connect through proxy

These are the hook — short, copy-pasteable, satisfying in under 2 minutes.

### Docs

The 44 existing wiki pages reorganized into a Starlight sidebar:

```
Getting Started
  ├── Installation
  ├── Quick Start: MySQL Single
  ├── Quick Start: MySQL Replication
  ├── Quick Start: PostgreSQL
  └── Quick Start: ProxySQL Integration

Core Concepts
  ├── Sandboxes
  ├── Versions & Flavors
  ├── Ports & Networking
  └── Environment Variables

Deploying
  ├── Single Sandbox
  ├── Multiple Sandboxes
  ├── Replication
  ├── Group Replication
  ├── Fan-In & All-Masters
  └── NDB Cluster

Providers
  ├── MySQL
  ├── PostgreSQL
  ├── ProxySQL
  └── Percona XtraDB Cluster

Managing Sandboxes
  ├── Starting & Stopping
  ├── Using Sandboxes
  ├── Customization
  ├── Database Users
  ├── Logs
  └── Deletion & Cleanup

Advanced
  ├── Concurrent Deployment
  ├── Importing Databases
  ├── Inter-Sandbox Replication
  ├── Cloning
  ├── Using as a Go Library
  └── Compiling from Source

Reference
  ├── CLI Commands
  ├── Configuration
  └── API Changelog
```

**Content strategy:** existing wiki markdown is kept mostly as-is. Navigation is restructured. Pages that don't fit are merged or dropped. Frontmatter is added/adjusted during the build copy step.

### Providers Page

Custom layout at `/providers` — the marketing angle for the provider architecture.

**Structure:**

1. **Intro** — dbdeployer's provider architecture, one CLI for multiple databases
2. **Comparison matrix:**

| | MySQL | PostgreSQL | ProxySQL |
|---|---|---|---|
| Single sandbox | ✓ | ✓ | ✓ |
| Multiple sandboxes | ✓ | ✓ | — |
| Replication | ✓ | ✓ (streaming) | — |
| Group replication | ✓ | — | — |
| ProxySQL wiring | ✓ | ✓ | — |
| Binary source | Tarballs | .deb extraction | System binary |

Note: MariaDB and Percona Server are MySQL-compatible flavors (same binary format, same provider) and are not listed as separate columns. The docs explain this under Providers > MySQL.

3. **Per-provider cards** — description, example command, link to docs
4. **"Coming Soon" teaser** — Orchestrator integration (from roadmap)

This is where ProxySQL gets introduced naturally — users browsing providers see the integration story.

### Blog

Content collection in `src/content/blog/`. Each post is a `.md` with frontmatter (title, date, author, tags, description).

**Blog index** at `/blog` — reverse-chronological, custom layout.

**Launch posts:**
1. "dbdeployer Under New Maintainership" — ProxySQL team story, what changed, roadmap
2. "PostgreSQL Support is Here" — Phase 3 announcement, examples

**Home integration:** latest 1-2 posts shown in "What's New" strip above footer.

## Docs Content Pipeline

Wiki pages are authored in `docs/wiki/` (close to the Go code). A build script (`website/scripts/copy-wiki.sh`) copies them into Starlight's content collection with transformations.

### Copy Script Responsibilities

The script (`copy-wiki.sh`) runs before `npm run build` and does:

1. **Copy files** from `docs/wiki/*.md` into `website/src/content/docs/<section>/` per the mapping table below
2. **Normalize filenames** — remove commas, double dots, convert to lowercase kebab-case
3. **Add Starlight frontmatter** — inject `title:` and `sidebar:` fields based on the mapping
4. **Rewrite links** — convert wiki-style links (`[text](other-page.md)`) to Starlight paths (`[text](/docs/<section>/other-page/)`)
5. **Strip wiki navigation** — remove `[[HOME]]`-style nav links (Starlight sidebar replaces these)
6. **Copy ProxySQL guide** — `docs/proxysql-guide.md` → `website/src/content/docs/providers/proxysql.md`

### Wiki Page Mapping

| Wiki File | Target Path | Sidebar Label |
|---|---|---|
| `installation.md` | `getting-started/installation` | Installation |
| *(new content)* | `getting-started/quickstart-mysql-single` | Quick Start: MySQL Single |
| *(new content)* | `getting-started/quickstart-mysql-replication` | Quick Start: MySQL Replication |
| *(new content)* | `getting-started/quickstart-postgresql` | Quick Start: PostgreSQL |
| *(new content)* | `getting-started/quickstart-proxysql` | Quick Start: ProxySQL Integration |
| `default-sandbox.md` | `concepts/sandboxes` | Sandboxes |
| `database-server-flavors.md` | `concepts/flavors` | Versions & Flavors |
| `ports-management.md` | `concepts/ports` | Ports & Networking |
| `../env_variables.md` | `concepts/environment-variables` | Environment Variables |
| `main-operations.md` | `deploying/single` | Single Sandbox |
| `multiple-sandboxes,-same-version-and-type.md` | `deploying/multiple` | Multiple Sandboxes |
| `replication-topologies.md` | `deploying/replication` | Replication |
| *(extract from replication-topologies.md)* | `deploying/group-replication` | Group Replication |
| *(extract from replication-topologies.md)* | `deploying/fan-in-all-masters` | Fan-In & All-Masters |
| *(extract from replication-topologies.md)* | `deploying/ndb-cluster` | NDB Cluster |
| `standard-and-non-standard-basedir-names.md` | `providers/mysql` | MySQL |
| *(new content)* | `providers/postgresql` | PostgreSQL |
| `../proxysql-guide.md` | `providers/proxysql` | ProxySQL |
| *(extract from replication-topologies.md)* | `providers/pxc` | Percona XtraDB Cluster |
| `skip-server-start.md` + `sandbox-management.md` | `managing/starting-stopping` | Starting & Stopping |
| `using-the-latest-sandbox.md` | `managing/using` | Using Sandboxes |
| `sandbox-customization.md` | `managing/customization` | Customization |
| `database-users.md` | `managing/users` | Database Users |
| `database-logs-management..md` | `managing/logs` | Logs |
| `sandbox-deletion.md` | `managing/deletion` | Deletion & Cleanup |
| `concurrent-deployment-and-deletion.md` | `advanced/concurrent` | Concurrent Deployment |
| `importing-databases-into-sandboxes.md` | `advanced/importing` | Importing Databases |
| `replication-between-sandboxes.md` | `advanced/inter-sandbox-replication` | Inter-Sandbox Replication |
| `cloning-databases.md` | `advanced/cloning` | Cloning |
| `using-dbdeployer-source-for-other-projects.md` | `advanced/go-library` | Using as a Go Library |
| `compiling-dbdeployer.md` | `advanced/compiling` | Compiling from Source |
| `command-line-completion.md` | `reference/cli-commands` | CLI Commands |
| `initializing-the-environment.md` | `reference/configuration` | Configuration |
| *(consolidated)* | `reference/api-changelog` | API Changelog |

### Dropped/Merged Pages

These wiki pages are NOT mapped to the sidebar (content merged into other pages or no longer relevant):

| Wiki File | Disposition |
|---|---|
| `Home.md` | Replaced by landing page |
| `do-not-edit.md` | Internal tooling note, drop |
| `generating-additional-documentation.md` | Internal tooling, drop |
| `semantic-versioning.md` | Merge into Reference > Configuration |
| `practical-examples.md` | Content absorbed into quickstart guides |
| `sandbox-macro-operations.md` | Merge into Managing > Using Sandboxes |
| `sandbox-upgrade.md` | Merge into Managing > Using Sandboxes |
| `dedicated-admin-address.md` | Merge into Deploying > Single Sandbox |
| `running-sysbench.md` | Merge into Advanced > Importing (or drop) |
| `mysql-document-store,-mysqlsh,-and-defaults..md` | Merge into Providers > MySQL |
| `installing-mysql-shell.md` | Merge into Providers > MySQL |
| `loading-sample-data-into-sandboxes.md` | Merge into Advanced > Importing |
| `using-dbdeployer-in-scripts.md` | Merge into Advanced > Go Library |
| `using-short-version-numbers.md` | Merge into Concepts > Versions & Flavors |
| `using-the-direct-path-to-the-expanded-tarball.md` | Merge into Concepts > Versions & Flavors |
| `getting-remote-tarballs.md` | Merge into Getting Started > Installation |
| `updating-dbdeployer.md` | Merge into Getting Started > Installation |
| `obtaining-sandbox-metadata.md` | Merge into Managing > Using Sandboxes |
| `exporting-dbdeployer-structure.md` | Merge into Reference > CLI Commands |
| `dbdeployer-operations-logging.md` | Merge into Managing > Logs |

### API Changelog Strategy

The 54 API version files (`docs/API/API-1.0.md` through `docs/API/1.68.md`) are **not** published individually. Instead:

- A single `reference/api-changelog.md` page is generated that consolidates the last 5 versions with full content
- Older versions link to the GitHub directory: "See [full API history on GitHub](https://github.com/ProxySQL/dbdeployer/tree/master/docs/API)"

This keeps the sidebar clean and avoids 54 pages of version diffs.

### Pipeline Summary

- Docs live near the code (developers edit `docs/wiki/`)
- The website automatically picks up changes
- No manual sync between repo and site

## Assets & Metadata

### SEO & Social

- **Favicon:** `public/favicon.svg` — simple dbdeployer logo/icon
- **OG image:** `public/og-image.png` — branded card (1200x630) with tagline, used as default `og:image`
- **Meta tags:** Starlight handles `<title>` and `<meta description>` from frontmatter for docs pages. Custom pages (landing, providers, blog) set their own `<meta>` tags in `<head>`
- **Sitemap:** Astro's `@astrojs/sitemap` integration generates `sitemap.xml` automatically

### Wiki Deprecation

After the website launches, add a notice to the top of the GitHub wiki `Home.md` (if the wiki is still accessible):

> "This wiki has moved to [proxysql.github.io/dbdeployer](https://proxysql.github.io/dbdeployer/docs/). These pages are no longer maintained."

The wiki pages in `docs/wiki/` remain in the repo as the source of truth — they're just served through the website now.

## Deployment

**Workflow:** `.github/workflows/deploy-website.yml`

Triggers:
- Push to `master` when `website/**` or `docs/wiki/**` change
- Manual `workflow_dispatch`

Steps:
1. Checkout repo
2. Setup Node.js 20 LTS (`actions/setup-node` with `node-version: '20'`)
3. `npm ci` in `website/`
4. Run copy script: `bash website/scripts/copy-wiki.sh` — transforms and copies `docs/wiki/*.md` into `website/src/content/docs/`
5. `npm run build`
6. Deploy `dist/` to `gh-pages` branch via `actions/deploy-pages`

**Site URL:** `proxysql.github.io/dbdeployer` (GitHub Pages default for org repos). Custom domain can be configured later.

**GitHub Pages config:** Settings → Pages → Source: GitHub Actions.

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
- **Hosting:** GitHub Pages
- **Deployment:** GitHub Actions → builds Astro → pushes to `gh-pages` branch

## Project Structure

```
website/
  astro.config.mjs
  package.json
  src/
    content/
      docs/           # Starlight docs (migrated wiki pages)
      blog/           # Blog posts as .md files
    pages/
      index.astro     # Landing page (custom, not Starlight)
      providers.astro # Providers comparison page
    components/       # Reusable Astro components (Hero, FeatureGrid, etc.)
    layouts/          # Custom layouts for landing/blog
    styles/           # Global CSS
  public/
    images/           # Screenshots, diagrams
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

Wiki pages are authored in `docs/wiki/` (close to the Go code). A build step copies them into Starlight's content collection:

1. GitHub Actions triggers on push to `master` when `website/**` or `docs/wiki/**` change
2. A script copies `docs/wiki/*.md` into `website/src/content/docs/`, mapping filenames to the sidebar structure and adding/adjusting Starlight frontmatter
3. `npm run build` generates the static site
4. Built output is deployed to `gh-pages` branch

This means:
- Docs live near the code (developers edit `docs/wiki/`)
- The website automatically picks up changes
- No manual sync between repo and site

## Deployment

**Workflow:** `.github/workflows/deploy-website.yml`

Triggers:
- Push to `master` when `website/**` or `docs/wiki/**` change
- Manual `workflow_dispatch`

Steps:
1. Checkout repo
2. Setup Node.js (LTS)
3. `npm ci` in `website/`
4. Run copy script: `docs/wiki/*.md` → `website/src/content/docs/` with frontmatter mapping
5. `npm run build`
6. Deploy `dist/` to `gh-pages` branch via `actions/deploy-pages`

**Site URL:** `proxysql.github.io/dbdeployer` (GitHub Pages default for org repos). Custom domain can be configured later.

**GitHub Pages config:** Settings → Pages → Source: GitHub Actions.

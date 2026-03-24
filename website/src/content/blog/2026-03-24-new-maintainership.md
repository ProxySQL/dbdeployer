---
title: "dbdeployer Under New Maintainership"
date: 2026-03-24
author: "Rene Cannao"
description: "The ProxySQL team takes over dbdeployer with modern MySQL support, a provider architecture, and PostgreSQL on the horizon."
tags: ["announcement", "roadmap"]
---

We're excited to announce that dbdeployer is now maintained by the ProxySQL team.

## What's Changed

Since taking over in March 2026, we've:

- **Modernized the stack** — Go 1.22+, refreshed dependencies, fixed all CVEs
- **Added MySQL 8.4 LTS and 9.x support** — full compatibility with modern MySQL
- **Built a provider architecture** — extensible system for deploying different database types
- **Integrated ProxySQL** — deploy read/write split stacks with `--with-proxysql`
- **Added PostgreSQL support** — streaming replication, deb-based binary management

## Why ProxySQL?

We use dbdeployer daily to test ProxySQL against every MySQL topology. When we took over maintainership, we saw an opportunity to make it useful for a much wider audience — anyone who needs local database sandboxes for development and testing.

## What's Next

- Orchestrator integration for failover testing
- More PostgreSQL topologies
- A proper website (you're looking at it!)

Follow the [GitHub repository](https://github.com/ProxySQL/dbdeployer) for updates.

---
title: "Admin Web UI"
description: "Visual dashboard for managing dbdeployer sandboxes"
---

`dbdeployer admin ui` launches a localhost web dashboard for managing all your deployed sandboxes visually.

## Quick Start

```bash
dbdeployer admin ui
```

A browser opens with a one-time authentication URL. The dashboard shows all your sandboxes with their status.

## Features

- **Sandbox cards** — each sandbox shown as a card with name, type, version, ports, and status badge (running/stopped)
- **Topology grouping** — replication sandboxes show primary + replica nodes together
- **Start/Stop/Destroy** — click buttons to manage sandbox lifecycle (no terminal needed)
- **Auto-refresh** — status updates every 5 seconds
- **Dark theme** — matches the dbdeployer website aesthetic

## Authentication

The admin UI uses one-time password (OTP) authentication:

1. `dbdeployer admin ui` prints a URL with a token to the terminal
2. Your browser opens to that URL automatically
3. The token is validated and a session cookie is set
4. The token is single-use — it can't be reused
5. Sessions expire after 1 hour

The server only binds to `127.0.0.1` (localhost) — it's never accessible from the network.

## Options

```bash
# Use a custom port (default: 9090)
dbdeployer admin ui --port 8080
```

## How It Works

- The server reads the sandbox catalog (`~/.dbdeployer/sandboxes.json`)
- Status is determined by running each sandbox's `status` script
- Start/stop/destroy actions execute the sandbox's own lifecycle scripts
- The UI uses HTMX for dynamic updates without page reloads
- All templates and assets are embedded in the Go binary — no external dependencies

## Security

- Localhost only (127.0.0.1)
- Single-use OTP authentication
- HttpOnly + SameSite cookies
- HTTP server timeouts (read 15s, write 30s, idle 60s)
- Path traversal protection on sandbox names

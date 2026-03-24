# Admin Web UI POC Design

**Date:** 2026-03-24
**Author:** Rene (ProxySQL)
**Status:** POC

## Goal

Prove that dbdeployer can be a platform, not just a CLI. A `dbdeployer admin` command launches a localhost web dashboard showing all deployed sandboxes with start/stop/destroy controls.

## Scope (POC only)

- Dashboard showing all sandboxes as cards grouped by topology
- Start/stop/destroy actions via the UI
- OTP authentication (CLI generates token, browser validates)
- Localhost only (127.0.0.1)
- Go templates + HTMX, embedded in binary

## NOT in scope (future)

- Deploy new sandboxes via UI
- Real-time log streaming
- Topology graph visualization
- Multi-user / remote access
- Persistent sessions

## Architecture

```
dbdeployer admin
  └─ starts HTTP server on 127.0.0.1:<port>
  └─ generates OTP, prints to terminal
  └─ opens browser to http://127.0.0.1:<port>/login?token=<otp>
  └─ serves embedded HTML templates via Go's html/template
  └─ HTMX handles dynamic actions (no page reload for start/stop/destroy)
  └─ API endpoints read sandbox catalog + execute lifecycle commands
```

### Authentication Flow

1. `dbdeployer admin` generates a random OTP (32-char hex)
2. Prints: `Admin UI: http://127.0.0.1:9090/login?token=<otp>`
3. Browser hits `/login?token=<otp>` → server validates → sets session cookie
4. Session cookie used for all subsequent requests
5. OTP is single-use (invalidated after first login)
6. Session expires when server stops (in-memory)

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/login` | Validate OTP, set session cookie, redirect to dashboard |
| GET | `/` | Dashboard (HTML) |
| GET | `/api/sandboxes` | JSON list of all sandboxes |
| POST | `/api/sandboxes/:name/start` | Start a sandbox |
| POST | `/api/sandboxes/:name/stop` | Stop a sandbox |
| POST | `/api/sandboxes/:name/destroy` | Destroy a sandbox (requires confirmation) |

### Dashboard Layout

**Header:** "dbdeployer admin" + sandbox count + server uptime

**Sandbox cards grouped by topology:**

```
┌─ Replication: rsandbox_8_4_4 ────────────────────────┐
│                                                        │
│  ┌─ master ─────────┐  ┌─ node1 ──────────┐          │
│  │ Port: 8404       │  │ Port: 8405       │          │
│  │ ● Running        │  │ ● Running        │          │
│  │ [Stop]           │  │ [Stop]           │          │
│  └──────────────────┘  └──────────────────┘          │
│                                                        │
│  ┌─ node2 ──────────┐  ┌─ proxysql ───────┐          │
│  │ Port: 8406       │  │ Port: 6032/6033  │          │
│  │ ● Running        │  │ ● Running        │          │
│  │ [Stop]           │  │ [Stop]           │          │
│  └──────────────────┘  └──────────────────┘          │
│                                                        │
│  [Stop All]  [Destroy] ──────────────────────────────│
└────────────────────────────────────────────────────────┘

┌─ Single: msb_8_4_4 ──────────────────────────────────┐
│  Port: 8404  │  ● Running  │  [Stop] [Destroy]       │
└────────────────────────────────────────────────────────┘
```

### Sandbox Data Source

Read from `~/.dbdeployer/sandboxes.json` (the existing sandbox catalog). Each entry has:
- Sandbox name and directory
- Type (single, multiple, replication, group, etc.)
- Ports
- Nodes (for multi-node topologies)

Status is determined by checking if the sandbox's PID file exists / process is running.

### Technology

- **Server:** Go `net/http` (stdlib, no framework)
- **Templates:** Go `html/template` with `//go:embed`
- **Interactivity:** HTMX (loaded from CDN or embedded)
- **Styling:** Inline CSS in the template (single file, dark theme matching the website)
- **Session:** In-memory map, cookie-based

## File Structure

```
cmd/admin.go              # Cobra command: dbdeployer admin
admin/
  server.go               # HTTP server, routes, middleware
  auth.go                 # OTP generation, session management
  handlers.go             # API handlers (list, start, stop, destroy)
  sandbox_status.go       # Read catalog, check process status
  templates/
    layout.html           # Base layout (head, nav, footer)
    dashboard.html         # Dashboard with sandbox cards
    login.html            # Login page (auto-submits with OTP)
    components/
      sandbox-card.html   # Single sandbox card partial
      topology-group.html # Topology group wrapper partial
  static/
    htmx.min.js           # HTMX library (embedded)
    style.css             # Dashboard styles
```

All templates and static files embedded via `//go:embed admin/templates/* admin/static/*`.

## Port Selection

Default: 9090. If busy, find next free port. Print the URL to terminal.
Flag: `--port` to override.

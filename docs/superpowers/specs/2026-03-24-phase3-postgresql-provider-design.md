# Phase 3 — PostgreSQL Provider Design

**Date:** 2026-03-24
**Author:** Rene (ProxySQL)
**Status:** Draft
**Prerequisite:** Phase 2b complete (provider interface, MySQL/ProxySQL providers)

## Context

dbdeployer's provider architecture (Phase 2) introduced a `Provider` interface with MySQL and ProxySQL implementations. Phase 3 validates that this architecture scales to a fundamentally different database system — PostgreSQL — where initialization, configuration, replication, and binary management all differ significantly from MySQL.

**Primary motivation:** Enable ProxySQL protocol compatibility testing against PostgreSQL backends, and prove the provider model generalizes beyond MySQL-family databases.

## Scope

- PostgreSQL provider: binary management (deb extraction), single sandbox, lifecycle
- Streaming replication topology
- Cross-database topology constraints and validation
- ProxySQL + PostgreSQL backend wiring
- Unit tests from day one; integration tests written but CI-gated as manual

## Provider Interface Changes

Two methods added to the `Provider` interface:

```go
type Provider interface {
    // ... existing methods ...

    // SupportedTopologies returns which topology types this provider can deploy.
    // The cmd layer validates against this before attempting deployment.
    SupportedTopologies() []string

    // CreateReplica creates a replica from a running primary instance.
    // Returns ErrNotSupported if the provider doesn't support replication.
    // Called by the topology layer after the primary is started.
    CreateReplica(primary SandboxInfo, config SandboxConfig) (*SandboxInfo, error)
}
```

**Per-provider topology support:**

| Provider   | Supported Topologies                                                       |
|------------|---------------------------------------------------------------------------|
| mysql      | single, multiple, replication, group, fan-in, all-masters, ndb, pxc      |
| proxysql   | single                                                                     |
| postgresql | single, multiple, replication                                              |

**MySQL provider** returns the full topology list from `SupportedTopologies()` — these topologies are served by the legacy `sandbox` package, not through the provider interface's `CreateSandbox`/`CreateReplica` methods. The topology list is accurate to what dbdeployer can deploy; it just flows through the old code path. `CreateReplica` returns `ErrNotSupported`.

**ProxySQL provider** returns `["single"]` and `ErrNotSupported` from `CreateReplica`.

**Binary resolution in `CreateReplica`:** The replica's `config.Version` is used to resolve binaries internally via `FindBinary(config.Version)`. This avoids needing to pass basedir through `SandboxInfo`.

### Cleanup on Failure

If `CreateSandbox` or `CreateReplica` fails partway through (e.g., initdb succeeds but config generation fails), the method cleans up its own sandbox directory before returning the error. The caller is not responsible for partial cleanup within a single sandbox.

For multi-node replication topologies, if replica N fails, the topology layer is responsible for stopping and destroying the primary and any previously created replicas. This matches the existing MySQL behavior where partial topology failures trigger full cleanup.

## Binary Management — Deb Extraction

PostgreSQL does not distribute pre-compiled tarballs. Binaries are extracted from `.deb` packages.

### Usage

```bash
# User downloads debs (familiar apt workflow)
apt-get download postgresql-16 postgresql-client-16

# dbdeployer extracts and lays out binaries
dbdeployer unpack --provider=postgresql postgresql-16_16.13.deb postgresql-client-16_16.13.deb
```

### Extraction Flow

1. Validate both debs are provided (server + client)
2. Extract each via `dpkg-deb -x` to a temp directory
3. Copy `usr/lib/postgresql/16/bin/` → `~/opt/postgresql/16.13/bin/`
4. Copy `usr/lib/postgresql/16/lib/` → `~/opt/postgresql/16.13/lib/`
5. Copy `usr/share/postgresql/16/` → `~/opt/postgresql/16.13/share/`
6. Validate required binaries exist: `postgres`, `initdb`, `pg_ctl`, `psql`, `pg_basebackup`
7. Clean up temp directory

### Version Detection

Extracted from deb filename pattern `postgresql-NN_X.Y-*`. Overridable via `--version=16.13`.

### Target Layout

```
~/opt/postgresql/16.13/
  bin/    (postgres, initdb, pg_ctl, psql, pg_basebackup, pg_dump, ...)
  lib/    (shared libraries)
  share/  (timezone data, extension SQL — required by initdb)
```

**Implementation:** `providers/postgresql/unpack.go`, called from `cmd/unpack.go` when `--provider=postgresql`.

## PostgreSQL Provider — Single Sandbox

### Registration

Same pattern as ProxySQL: `Register()` called from `cmd/root.go` init.

### Port Allocation

`DefaultPorts()` returns `{BasePort: 15000, PortsPerInstance: 1}`.

Version-to-port formula: `BasePort + major * 100 + minor`. Examples:
- `16.13` → `15000 + 1600 + 13` = `16613`
- `16.3` → `15000 + 1600 + 3` = `16603`
- `17.1` → `15000 + 1700 + 1` = `16701`
- `17.10` → `15000 + 1700 + 10` = `16710`

Single port per instance (PostgreSQL uses one port for all connections).

### Version Validation

`ValidateVersion()` accepts exactly `major.minor` format where both parts are integers. Major must be >= 12 (oldest supported PostgreSQL with streaming replication via `pg_basebackup -R`). Three-part versions like `16.13.1` are rejected (PostgreSQL does not use them).

### FindBinary

Looks in `~/opt/postgresql/<version>/bin/postgres`. Provider determines base path (`~/opt/postgresql/`); `--basedir` overrides for custom locations.

### CreateSandbox Flow

1. **Create log directory:** `mkdir -p <sandbox>/data/log`

2. **Init database:**
   ```bash
   initdb -D <sandbox>/data --auth=trust --username=postgres
   ```
   Note: `initdb` locates `share/` data relative to its own binary path (`../share/`). Since the extraction layout places `share/` as a sibling of `bin/`, no `-L` flag is needed. If the layout ever changes, `-L <basedir>/share` can be added as a fallback.

3. **Generate `postgresql.conf`:**
   ```
   port = <assigned_port>
   listen_addresses = '127.0.0.1'
   unix_socket_directories = '<sandbox>/data'
   logging_collector = on
   log_directory = '<sandbox>/data/log'
   ```

4. **Generate `pg_hba.conf`** (overwrite initdb default):
   ```
   local   all   all                trust
   host    all   all   127.0.0.1/32 trust
   host    all   all   ::1/128      trust
   ```

5. **Write lifecycle scripts** (inline generation, like ProxySQL):
   - `start` — `pg_ctl -D <data> -l <sandbox>/postgresql.log start`
   - `stop` — `pg_ctl -D <data> stop -m fast`
   - `status` — `pg_ctl -D <data> status`
   - `restart` — `pg_ctl -D <data> -l <sandbox>/postgresql.log restart`
   - `use` — `psql -h 127.0.0.1 -p <port> -U postgres`
   - `clear` — stop + remove data directory + re-init

6. **Set environment in all scripts:**
   - `LD_LIBRARY_PATH=<basedir>/lib/` (extracted debs need this for shared libraries)
   - Unset `PGDATA`, `PGPORT`, `PGHOST`, `PGUSER`, `PGDATABASE` to prevent environment contamination from the user's shell

7. **Return `SandboxInfo`** with dir, port. `Socket` field is left empty (lifecycle scripts use TCP via `127.0.0.1`, matching the ProxySQL provider pattern). The unix socket exists at `<sandbox>/data/.s.PGSQL.<port>` but is not the primary connection method.

### Multiple Topology

`dbdeployer deploy multiple 16.13 --provider=postgresql` creates N independent PostgreSQL instances using `CreateSandbox` with sequential port allocation. No additional configuration beyond what single provides — each instance is standalone with no replication relationship.

## PostgreSQL Replication

### CreateReplica Flow

1. **No `initdb`** — replica data comes from the running primary via `pg_basebackup`:
   ```bash
   pg_basebackup -h 127.0.0.1 -p <primary_port> -U postgres -D <replica>/data -Fp -Xs -R
   ```
   - `-Fp` = plain format
   - `-Xs` = stream WAL during backup
   - `-R` = auto-create `standby.signal` + write `primary_conninfo` to `postgresql.auto.conf`

2. **Modify replica's `postgresql.conf`:**
   - Change `port` to replica's assigned port
   - Change `unix_socket_directories` to replica's sandbox dir

3. **Write lifecycle scripts** — same as single sandbox with replica's port

4. **Start replica** — `pg_ctl -D <data> -l <sandbox>/postgresql.log start`

### Primary-Side Configuration

When replication is intended (`config.Options["replication"] = "true"`), `CreateSandbox` adds:

**postgresql.conf:**
```
wal_level = replica
max_wal_senders = 10
hot_standby = on
```

**pg_hba.conf:**
```
host    replication    all    127.0.0.1/32    trust
```

### Topology Layer Flow

For `dbdeployer deploy replication 16.13 --provider=postgresql`:

1. `CreateSandbox()` for primary with replication options
2. `StartSandbox()` for primary — **must be running before replicas**
3. For each replica: `CreateReplica(primaryInfo, replicaConfig)` — **sequential, not concurrent**
4. Each replica starts automatically as part of `CreateReplica`

### Monitoring Scripts

Generated in the topology directory:

**`check_replication`** — connects to primary, shows connected replicas:
```bash
psql -h 127.0.0.1 -p <primary_port> -U postgres -c \
  "SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn FROM pg_stat_replication;"
```

**`check_recovery`** — connects to each replica, verifies standby status:
```bash
# For each replica:
psql -h 127.0.0.1 -p <replica_port> -U postgres -c "SELECT pg_is_in_recovery();"
```

## ProxySQL + PostgreSQL Wiring

### Triggering

```bash
dbdeployer deploy replication 16.13 --provider=postgresql --with-proxysql
```

### Config Generation

ProxySQL supports PostgreSQL backends natively. The backend provider type is passed via:
```go
config.Options["backend_provider"] = "postgresql"
```

The ProxySQL config generator (`providers/proxysql/config.go`) branches:

| Backend Provider | Config Blocks                                          |
|------------------|-------------------------------------------------------|
| mysql (default)  | `mysql_servers`, `mysql_users`, `mysql_variables`     |
| postgresql       | `pgsql_servers`, `pgsql_users`, `pgsql_variables`     |

### End-to-End Flow

1. Deploy PostgreSQL primary + replicas (streaming replication)
2. Deploy ProxySQL with `pgsql_servers` pointing to primary (HG 0) + replicas (HG 1)
3. Generate `use_proxy` script: `psql -h 127.0.0.1 -p <proxysql_pgsql_port> -U postgres`

### Port Allocation

ProxySQL admin port stays on its usual range (6032+). The frontend port uses the next consecutive port, same as today. The ProxySQL `ProxySQLConfig` struct's `MySQLPort` field is reused for the frontend listener port regardless of backend type — the field name is a misnomer but changing it would break the MySQL path. The `use_proxy` script uses `psql` instead of `mysql` when `backend_provider` is `postgresql`.

## Cross-Database Topology Constraints

### Topology Validation

Cmd layer validates provider supports the requested topology before any sandbox creation:

```
$ dbdeployer deploy group 16.13 --provider=postgresql
Error: provider "postgresql" does not support topology "group"
Supported topologies: single, multiple, replication
```

### Flavor Validation

`--flavor` is MySQL-specific. Rejected when `--provider` is not `mysql`:

```
$ dbdeployer deploy single 16.13 --provider=postgresql --flavor=ndb
Error: --flavor is only valid with --provider=mysql
```

### Cross-Provider Wiring Validation

Compatibility map determines which addons work with which providers:

```go
var compatibleAddons = map[string][]string{
    "proxysql": {"mysql", "postgresql"},
    // future: "orchestrator": {"mysql"},
}
```

```
$ dbdeployer deploy single 16.13 --provider=postgresql --with-orchestrator
Error: --with-orchestrator is not compatible with provider "postgresql"
```

## Testing Strategy

### Unit Tests (no binaries needed)

- `providers/postgresql/postgresql_test.go` — `ValidateVersion()`, `DefaultPorts()`, `SupportedTopologies()`, port calculation, config generation (postgresql.conf, pg_hba.conf), script generation
- `providers/postgresql/unpack_test.go` — deb filename parsing, version extraction, required binary validation
- `providers/proxysql/config_test.go` — extend for PostgreSQL backend config (`pgsql_servers`/`pgsql_users`)
- `providers/provider_test.go` — extend for topology validation, flavor rejection, cross-provider compatibility
- Cmd-level tests — `--provider=postgresql --flavor=ndb` errors, unsupported topologies error

### Integration Tests (`//go:build integration`)

`providers/postgresql/integration_test.go`:
- Single sandbox: initdb → start → connect via psql → stop → destroy
- Replication: primary + 2 replicas → verify `pg_stat_replication` shows 2 senders → verify `pg_is_in_recovery() = true`
- With ProxySQL: replication + proxysql → connect through ProxySQL → verify routing
- Deb extraction: unpack real .deb files → verify binary layout

### CI Follow-Up (tracked as GitHub issues)

1. Add PostgreSQL deb caching to CI pipeline
2. Add PostgreSQL integration tests to CI matrix
3. Nightly topology tests for PostgreSQL replication

Integration tests run locally until CI is set up.

## File Structure

```
providers/postgresql/
  postgresql.go          # Provider implementation (CreateSandbox, CreateReplica, lifecycle)
  unpack.go              # Deb extraction logic
  config.go              # postgresql.conf and pg_hba.conf generation
  postgresql_test.go     # Unit tests
  unpack_test.go         # Deb extraction unit tests
  integration_test.go    # Integration tests (build-tagged)
```

Modifications to existing files:
- `providers/provider.go` — add `SupportedTopologies()`, `CreateReplica()` to interface
- `providers/mysql/mysql.go` — implement new interface methods (return full topology list, ErrNotSupported for CreateReplica)
- `providers/proxysql/proxysql.go` — implement new interface methods
- `providers/proxysql/config.go` — PostgreSQL backend config generation
- `cmd/root.go` — register PostgreSQL provider
- `cmd/single.go`, `cmd/multiple.go`, `cmd/replication.go` — `--provider` flag, topology validation
- `cmd/unpack.go` — `--provider` flag for deb extraction
- `globals/globals.go` — PostgreSQL constants, flag labels

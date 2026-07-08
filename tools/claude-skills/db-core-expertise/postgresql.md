# PostgreSQL Notes

- `dbdeployer` expects user-space PostgreSQL binaries laid out as `bin/`, `lib/`, and `share/`.
- Debian and apt extraction plus share-dir wiring are common failure points.
- Validate `initdb` share paths, stop/start scripts, socket/config paths, and primary/replica setup.
- Edge cases:
  - wrong `-L` share dir for `initdb`
  - missing timezone or extension files
  - stale `postmaster.pid`
  - replica recovery config drift
- Good validation:
  - `~/sandboxes/pg_sandbox_*/use -c "SELECT version();"`
  - `bash ~/sandboxes/postgresql_repl_*/check_replication`
  - write on primary, read on replicas

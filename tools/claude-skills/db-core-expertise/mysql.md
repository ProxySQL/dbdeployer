# MySQL Notes

- `dbdeployer` commonly manages tarball-based MySQL layouts under `~/opt/mysql/<version>`.
- Watch for version differences across 8.0, 8.4, and 9.x.
- Verify defaults that changed across releases: auth plugin, mysqlx behavior, packaging names, startup scripts, and server flags.
- Edge cases:
  - missing shared libs on Linux
  - stale socket files
  - port collisions across mysql/mysqlx/admin ports
  - replication role ordering
- Good validation:
  - `~/sandboxes/.../use -e "SELECT VERSION();"`
  - `~/sandboxes/rsandbox_*/check_slaves`
  - `~/sandboxes/rsandbox_*/test_replication`

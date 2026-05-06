# ProxySQL Notes

- Track the admin and mysql listener pair together.
- Distinguish standalone deployment from topology-attached deployment.
- Validate backend registration, credentials, hostgroup wiring, and start/stop scripts.
- Edge cases:
  - admin port collision with listener pair
  - binary present but runtime dirs missing
  - backend auth mismatch
  - PostgreSQL proxy support gaps or work-in-progress behavior
- Good validation:
  - `~/sandboxes/*/proxysql/status`
  - `~/sandboxes/*/proxysql/use -e "SELECT * FROM mysql_servers;"`
  - `~/sandboxes/*/proxysql/use_proxy -e "SELECT 1;"`

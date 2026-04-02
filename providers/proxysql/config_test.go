package proxysql

import (
	"strings"
	"testing"
)

func TestGenerateConfigBasic(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost: "127.0.0.1", AdminPort: 6032,
		AdminUser: "admin", AdminPassword: "admin",
		MySQLPort: 6033, DataDir: "/tmp/test",
		MonitorUser: "msandbox", MonitorPass: "msandbox",
	}
	result := GenerateConfig(cfg)
	checks := []string{
		`admin_credentials="admin:admin"`,
		`interfaces="127.0.0.1:6033"`,
		`monitor_username="msandbox"`,
		`mysql_ifaces="127.0.0.1:6032"`,
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("missing %q in config output", check)
		}
	}
}

func TestGenerateConfigWithBackends(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost: "127.0.0.1", AdminPort: 6032,
		AdminUser: "admin", AdminPassword: "admin",
		MySQLPort: 6033, DataDir: "/tmp/test",
		MonitorUser: "msandbox", MonitorPass: "msandbox",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 100},
			{Host: "127.0.0.1", Port: 3307, Hostgroup: 1, MaxConns: 100},
		},
	}
	result := GenerateConfig(cfg)
	if !strings.Contains(result, "mysql_servers=") {
		t.Error("missing mysql_servers section")
	}
	if !strings.Contains(result, "port=3306") {
		t.Error("missing first backend")
	}
	if !strings.Contains(result, "hostgroup=1") {
		t.Error("missing reader hostgroup")
	}
}

func TestGenerateConfigMySQL(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:     "127.0.0.1",
		AdminPort:     6032,
		AdminUser:     "admin",
		AdminPassword: "admin",
		MySQLPort:     6033,
		DataDir:       "/tmp/proxysql/data",
		MonitorUser:   "msandbox",
		MonitorPass:   "msandbox",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "mysql_servers") {
		t.Error("expected mysql_servers block")
	}
	if !strings.Contains(config, "mysql_variables") {
		t.Error("expected mysql_variables block")
	}
	if !strings.Contains(config, "mysql_users") {
		t.Error("expected mysql_users block")
	}
}

func TestGenerateConfigPostgreSQL(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost:       "127.0.0.1",
		AdminPort:       6032,
		AdminUser:       "admin",
		AdminPassword:   "admin",
		MySQLPort:       6033,
		DataDir:         "/tmp/proxysql/data",
		MonitorUser:     "postgres",
		MonitorPass:     "postgres",
		BackendProvider: "postgresql",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 16613, Hostgroup: 0, MaxConns: 200},
			{Host: "127.0.0.1", Port: 16614, Hostgroup: 1, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "pgsql_servers") {
		t.Error("expected pgsql_servers block")
	}
	if !strings.Contains(config, "pgsql_users") {
		t.Error("expected pgsql_users block")
	}
	if !strings.Contains(config, "pgsql_variables") {
		t.Error("expected pgsql_variables block")
	}
	if strings.Contains(config, "mysql_servers") {
		t.Error("should not contain mysql_servers for postgresql backend")
	}
	if strings.Contains(config, "mysql_variables") {
		t.Error("should not contain mysql_variables for postgresql backend")
	}
}

func TestGenerateConfigGRHostgroups(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost: "127.0.0.1", AdminPort: 6032,
		AdminUser: "admin", AdminPassword: "admin",
		MySQLPort: 6033, DataDir: "/tmp/test",
		MonitorUser: "msandbox", MonitorPass: "msandbox",
		Topology: "innodb-cluster",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 200},
			{Host: "127.0.0.1", Port: 3307, Hostgroup: 1, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "mysql_group_replication_hostgroups") {
		t.Error("expected mysql_group_replication_hostgroups for innodb-cluster topology")
	}
	if !strings.Contains(config, "writer_hostgroup=0") {
		t.Error("expected writer_hostgroup=0")
	}
	if !strings.Contains(config, "reader_hostgroup=1") {
		t.Error("expected reader_hostgroup=1")
	}
	if !strings.Contains(config, "writer_is_also_reader=1") {
		t.Error("expected writer_is_also_reader=1")
	}
}

func TestGenerateConfigNoGRHostgroupsForReplication(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost: "127.0.0.1", AdminPort: 6032,
		AdminUser: "admin", AdminPassword: "admin",
		MySQLPort: 6033, DataDir: "/tmp/test",
		MonitorUser: "msandbox", MonitorPass: "msandbox",
		Topology: "replication",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if strings.Contains(config, "mysql_group_replication_hostgroups") {
		t.Error("should not contain mysql_group_replication_hostgroups for standard replication")
	}
}

func TestGenerateConfigGRHostgroupsForGroup(t *testing.T) {
	cfg := ProxySQLConfig{
		AdminHost: "127.0.0.1", AdminPort: 6032,
		AdminUser: "admin", AdminPassword: "admin",
		MySQLPort: 6033, DataDir: "/tmp/test",
		MonitorUser: "msandbox", MonitorPass: "msandbox",
		Topology: "group",
		Backends: []BackendServer{
			{Host: "127.0.0.1", Port: 3306, Hostgroup: 0, MaxConns: 200},
		},
	}
	config := GenerateConfig(cfg)
	if !strings.Contains(config, "mysql_group_replication_hostgroups") {
		t.Error("expected mysql_group_replication_hostgroups for group topology")
	}
}

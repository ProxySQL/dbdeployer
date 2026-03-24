package postgresql

import (
	"strings"
	"testing"
)

func TestGeneratePostgresqlConf(t *testing.T) {
	conf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:            5433,
		ListenAddresses: "127.0.0.1",
		UnixSocketDir:   "/tmp/sandbox/data",
		LogDir:          "/tmp/sandbox/data/log",
		Replication:     false,
	})
	if !strings.Contains(conf, "port = 5433") {
		t.Error("missing port setting")
	}
	if !strings.Contains(conf, "listen_addresses = '127.0.0.1'") {
		t.Error("missing listen_addresses")
	}
	if !strings.Contains(conf, "unix_socket_directories = '/tmp/sandbox/data'") {
		t.Error("missing unix_socket_directories")
	}
	if !strings.Contains(conf, "logging_collector = on") {
		t.Error("missing logging_collector")
	}
	if strings.Contains(conf, "wal_level") {
		t.Error("should not contain wal_level when replication is false")
	}
}

func TestGeneratePostgresqlConfWithReplication(t *testing.T) {
	conf := GeneratePostgresqlConf(PostgresqlConfOptions{
		Port:            5433,
		ListenAddresses: "127.0.0.1",
		UnixSocketDir:   "/tmp/sandbox/data",
		LogDir:          "/tmp/sandbox/data/log",
		Replication:     true,
	})
	if !strings.Contains(conf, "wal_level = replica") {
		t.Error("missing wal_level = replica")
	}
	if !strings.Contains(conf, "max_wal_senders = 10") {
		t.Error("missing max_wal_senders")
	}
	if !strings.Contains(conf, "hot_standby = on") {
		t.Error("missing hot_standby")
	}
}

func TestGeneratePgHbaConf(t *testing.T) {
	conf := GeneratePgHbaConf(false)
	if !strings.Contains(conf, "local   all") {
		t.Error("missing local all entry")
	}
	if !strings.Contains(conf, "host    all") {
		t.Error("missing host all entry")
	}
	if strings.Contains(conf, "replication") {
		t.Error("should not contain replication when replication is false")
	}
}

func TestGeneratePgHbaConfWithReplication(t *testing.T) {
	conf := GeneratePgHbaConf(true)
	if !strings.Contains(conf, "host    replication") {
		t.Error("missing replication entry")
	}
}

package postgresql

import (
	"fmt"
	"strings"
)

type PostgresqlConfOptions struct {
	Port            int
	ListenAddresses string
	UnixSocketDir   string
	LogDir          string
	Replication     bool
}

func GeneratePostgresqlConf(opts PostgresqlConfOptions) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("port = %d\n", opts.Port))
	b.WriteString(fmt.Sprintf("listen_addresses = '%s'\n", opts.ListenAddresses))
	b.WriteString(fmt.Sprintf("unix_socket_directories = '%s'\n", opts.UnixSocketDir))
	b.WriteString("logging_collector = on\n")
	b.WriteString(fmt.Sprintf("log_directory = '%s'\n", opts.LogDir))

	if opts.Replication {
		b.WriteString("\n# Replication settings\n")
		b.WriteString("wal_level = replica\n")
		b.WriteString("max_wal_senders = 10\n")
		b.WriteString("hot_standby = on\n")
	}

	return b.String()
}

func GeneratePgHbaConf(replication bool) string {
	var b strings.Builder
	b.WriteString("# TYPE  DATABASE  USER  ADDRESS       METHOD\n")
	b.WriteString("local   all       all                 trust\n")
	b.WriteString("host    all       all   127.0.0.1/32  trust\n")
	b.WriteString("host    all       all   ::1/128       trust\n")

	if replication {
		b.WriteString("host    replication  all  127.0.0.1/32  trust\n")
	}

	return b.String()
}

package proxysql

import (
	"fmt"
	"strings"
)

type BackendServer struct {
	Host      string
	Port      int
	Hostgroup int
	MaxConns  int
}

type ProxySQLConfig struct {
	AdminHost       string
	AdminPort       int
	AdminUser       string
	AdminPassword   string
	MySQLPort       int
	DataDir         string
	Backends        []BackendServer
	MonitorUser     string
	MonitorPass     string
	BackendProvider string
}

func GenerateConfig(cfg ProxySQLConfig) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("datadir=\"%s\"\n\n", cfg.DataDir))

	b.WriteString("admin_variables=\n{\n")
	b.WriteString(fmt.Sprintf("    admin_credentials=\"%s:%s\"\n", cfg.AdminUser, cfg.AdminPassword))
	b.WriteString(fmt.Sprintf("    mysql_ifaces=\"%s:%d\"\n", cfg.AdminHost, cfg.AdminPort))
	b.WriteString("}\n\n")

	isPgsql := cfg.BackendProvider == "postgresql"

	if isPgsql {
		b.WriteString("pgsql_variables=\n{\n")
		b.WriteString(fmt.Sprintf("    interfaces=\"%s:%d\"\n", cfg.AdminHost, cfg.MySQLPort))
		b.WriteString(fmt.Sprintf("    monitor_username=\"%s\"\n", cfg.MonitorUser))
		b.WriteString(fmt.Sprintf("    monitor_password=\"%s\"\n", cfg.MonitorPass))
		b.WriteString("}\n\n")
	} else {
		b.WriteString("mysql_variables=\n{\n")
		b.WriteString(fmt.Sprintf("    interfaces=\"%s:%d\"\n", cfg.AdminHost, cfg.MySQLPort))
		b.WriteString(fmt.Sprintf("    monitor_username=\"%s\"\n", cfg.MonitorUser))
		b.WriteString(fmt.Sprintf("    monitor_password=\"%s\"\n", cfg.MonitorPass))
		b.WriteString("    monitor_connect_interval=2000\n")
		b.WriteString("    monitor_ping_interval=2000\n")
		b.WriteString("}\n\n")
	}

	serversKey := "mysql_servers"
	usersKey := "mysql_users"
	if isPgsql {
		serversKey = "pgsql_servers"
		usersKey = "pgsql_users"
	}

	if len(cfg.Backends) > 0 {
		b.WriteString(fmt.Sprintf("%s=\n(\n", serversKey))
		for i, srv := range cfg.Backends {
			b.WriteString("    {\n")
			b.WriteString(fmt.Sprintf("        address=\"%s\"\n", srv.Host))
			b.WriteString(fmt.Sprintf("        port=%d\n", srv.Port))
			b.WriteString(fmt.Sprintf("        hostgroup=%d\n", srv.Hostgroup))
			maxConns := srv.MaxConns
			if maxConns == 0 {
				maxConns = 200
			}
			b.WriteString(fmt.Sprintf("        max_connections=%d\n", maxConns))
			b.WriteString("    }")
			if i < len(cfg.Backends)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(")\n\n")
	}

	b.WriteString(fmt.Sprintf("%s=\n(\n", usersKey))
	b.WriteString("    {\n")
	b.WriteString(fmt.Sprintf("        username=\"%s\"\n", cfg.MonitorUser))
	b.WriteString(fmt.Sprintf("        password=\"%s\"\n", cfg.MonitorPass))
	b.WriteString("        default_hostgroup=0\n")
	b.WriteString("    }\n")
	b.WriteString(")\n")

	return b.String()
}

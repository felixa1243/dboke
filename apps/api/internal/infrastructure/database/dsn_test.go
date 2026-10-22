package database

import (
	"testing"

	"dboke-api/internal/core/domain"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

var trickyConfig = domain.DBConnectionConfig{
	Host:     "db.example.com",
	Port:     5432,
	User:     "app@user",
	Password: "p@ss:w/rd?x=1#frag%",
	Database: "my db",
}

func TestPostgresDSN_EscapesCredentials(t *testing.T) {
	cfg, err := pgx.ParseConfig(postgresDSN(trickyConfig))
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.Host != trickyConfig.Host || cfg.Port != uint16(trickyConfig.Port) {
		t.Errorf("host/port = %s:%d", cfg.Host, cfg.Port)
	}
	if cfg.User != trickyConfig.User || cfg.Password != trickyConfig.Password || cfg.Database != trickyConfig.Database {
		t.Errorf("parsed user=%q password=%q db=%q", cfg.User, cfg.Password, cfg.Database)
	}
}

func TestMySQLDSN_EscapesCredentials(t *testing.T) {
	cfg, err := mysql.ParseDSN(mysqlDSN(trickyConfig))
	if err != nil {
		t.Fatalf("ParseDSN: %v", err)
	}
	if cfg.Addr != "db.example.com:5432" || cfg.User != trickyConfig.User || cfg.Passwd != trickyConfig.Password || cfg.DBName != trickyConfig.Database {
		t.Errorf("parsed addr=%q user=%q password=%q db=%q", cfg.Addr, cfg.User, cfg.Passwd, cfg.DBName)
	}
	if !cfg.ParseTime {
		t.Error("parseTime should be enabled")
	}
}

func TestMongoClientOptions(t *testing.T) {
	opts := mongoClientOptions(trickyConfig)
	if len(opts.Hosts) != 1 || opts.Hosts[0] != "db.example.com:5432" {
		t.Errorf("hosts = %v", opts.Hosts)
	}
	if opts.Auth == nil || opts.Auth.Username != trickyConfig.User || opts.Auth.Password != trickyConfig.Password {
		t.Errorf("auth = %+v", opts.Auth)
	}
	if mongoClientOptions(domain.DBConnectionConfig{Host: "h", Port: 1}).Auth != nil {
		t.Error("no credentials should be set when user is empty")
	}
}

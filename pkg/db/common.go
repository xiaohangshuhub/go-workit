package db

import (
	"time"

	"gorm.io/gorm"
)

const (
	MySQLDefaultDns     = "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local&allowPublicKeyRetrieval=true"
	PostgresDefaultDns  = "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	SQLServerDefaultDsn = "sqlserver://sa:your_password@localhost:1433?database=master"
	SQLiteDefaultDsn    = "test.db"
	MaxOpenConns        = 10
	MaxIdleConns        = 5
	ConnMaxLifetime     = 10 * time.Minute
)

type DatabaseConfig struct {
	LogLevel        string        `mapstructure:"log_level"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
	DryRun          bool          `mapstructure:"dry_run"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	*gorm.Config
}

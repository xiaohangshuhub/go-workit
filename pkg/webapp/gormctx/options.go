package gormctx

import (
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// Options 数据库选项
type Options struct {
	container   []fx.Option         // 持有容器引用
	databaseMap map[string]struct{} // 数据库实例名称集合
}

// NewOptions 创建数据库选项
func NewOptions() *Options {
	return &Options{
		container:   make([]fx.Option, 0),
		databaseMap: make(map[string]struct{}),
	}
}

// UseMySQL 使用 MySQL 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseMySQL(instanceName string, fn func(*gormx.MySQLConfigOptions)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &gormx.MySQLConfigOptions{
		DatabaseConfig: gormx.DatabaseConfig{
			MaxOpenConns:    gormx.MaxOpenConns,
			MaxIdleConns:    gormx.MaxIdleConns,
			ConnMaxLifetime: gormx.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		MySQLCfg: mysql.Config{},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return gormx.NewMysqlDB(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return gormx.NewMysqlDB(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UsePostgresSQL 使用 Postgres 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UsePostgresSQL(instanceName string, fn func(*gormx.PostgresConfigOptions)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &gormx.PostgresConfigOptions{
		DatabaseConfig: gormx.DatabaseConfig{
			MaxOpenConns:    gormx.MaxOpenConns,
			MaxIdleConns:    gormx.MaxIdleConns,
			ConnMaxLifetime: gormx.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		PgSQLCfg: postgres.Config{},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return gormx.NewPostgresDB(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return gormx.NewPostgresDB(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UseSQLServer 使用 UseSQLServer 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseSQLServer(instanceName string, fn func(*gormx.SQLServerConfigOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &gormx.SQLServerConfigOptions{
		DatabaseConfig: gormx.DatabaseConfig{
			MaxOpenConns:    gormx.MaxOpenConns,
			MaxIdleConns:    gormx.MaxIdleConns,
			ConnMaxLifetime: gormx.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		SQLServerCfg: sqlserver.Config{},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return gormx.NewSQLServer(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return gormx.NewSQLServer(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UseSQLServer 使用 UseSQLite 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseSQLite(instanceName string, fn func(*gormx.SQLiteConfigOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &gormx.SQLiteConfigOptions{
		DatabaseConfig: gormx.DatabaseConfig{
			MaxOpenConns:    gormx.MaxOpenConns,
			MaxIdleConns:    gormx.MaxIdleConns,
			ConnMaxLifetime: gormx.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		SQLiteCfg: sqlite.Config{},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return gormx.NewSQLite(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return gormx.NewSQLite(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

func (o *Options) Container() []fx.Option {
	return o.container
}

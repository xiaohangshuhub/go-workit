package dbctx

import (
	"github.com/xiaohangshuhub/go-workit/pkg/db"
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
func (d *Options) UseMySQL(instanceName string, fn func(cfg *db.MySQLConfigOptions)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &db.MySQLConfigOptions{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		MySQLCfg: mysql.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return db.NewMysqlDB(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return db.NewMysqlDB(lc, cfg, logger)
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
func (d *Options) UsePostgresSQL(instanceName string, fn func(cfg *db.PostgresConfigOptions)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &db.PostgresConfigOptions{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		PgSQLCfg: postgres.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return db.NewPostgresDB(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return db.NewPostgresDB(lc, cfg, logger)
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
func (d *Options) UseSQLServer(instanceName string, fn func(cfg *db.SQLServerConfigOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &db.SQLServerConfigOptions{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		SQLServerCfg: sqlserver.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return db.NewSQLServer(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return db.NewSQLServer(lc, cfg, logger)
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
func (d *Options) UseSQLite(instanceName string, fn func(cfg *db.SQLiteConfigOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &db.SQLiteConfigOptions{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		SQLiteCfg: sqlite.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return db.NewSQLite(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return db.NewSQLite(lc, cfg, logger)
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

package dbcontext

import (
	"github.com/xiaohangshuhub/go-workit/pkg/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Options 数据库选项
type Options struct {
	Container   []fx.Option         // 持有容器引用
	databaseMap map[string]struct{} // 数据库实例名称集合
}

// NewOptions 创建数据库选项
func NewOptions() *Options {
	return &Options{
		Container:   make([]fx.Option, 0),
		databaseMap: make(map[string]struct{}),
	}
}

// UseMySQL 注册 MySQL 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseMySQL(instanceName string, fn func(cfg *database.MySQLConfigOptions)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &database.MySQLConfigOptions{
		DatabaseConfig: database.DatabaseConfig{
			MaxOpenConns:    database.MaxOpenConns,
			MaxIdleConns:    database.MaxIdleConns,
			ConnMaxLifetime: database.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		MySQLCfg: mysql.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.Container = append(d.Container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return database.NewMysqlDB(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.Container = append(d.Container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return database.NewMysqlDB(lc, cfg, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UsePostgresSQL 注册 Postgres 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UsePostgresSQL(instanceName string, fn func(cfg *database.PostgresConfig)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &database.PostgresConfig{
		DatabaseConfig: database.DatabaseConfig{
			MaxOpenConns:    database.MaxOpenConns,
			MaxIdleConns:    database.MaxIdleConns,
			ConnMaxLifetime: database.ConnMaxLifetime,
			Config:          &gorm.Config{},
		},

		PgSQLCfg: postgres.Config{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.Container = append(d.Container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return database.NewPostgresDB(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.Container = append(d.Container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
						return database.NewPostgresDB(lc, cfg, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

func (d *Options) UseSQLServer(instanceName string, fn func(cfg *database.DatabaseConfig)) *Options {
	return d
}

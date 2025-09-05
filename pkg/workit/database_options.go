package workit

import (
	"time"

	"github.com/xiaohangshuhub/go-workit/pkg/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DatabaseOptions 数据库选项
type DatabaseOptions struct {
	container   []fx.Option         // 持有容器引用
	databaseMap map[string]struct{} // 数据库实例名称集合
}

// NewDatabaseOptions 创建数据库选项
func newDatabaseOptions() *DatabaseOptions {
	return &DatabaseOptions{
		container:   make([]fx.Option, 0),
		databaseMap: make(map[string]struct{}),
	}
}

// UseMySQL 注册 MySQL 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *DatabaseOptions) UseMySQL(instanceName string, fn func(cfg *database.MysqlConfigOptions)) *DatabaseOptions {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &database.MysqlConfigOptions{
		CommonDatabaseConfig: database.CommonDatabaseConfig{
			DSN:             "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
			LogLevel:        "debug",
			SlowThreshold:   100 * time.Millisecond,
			DryRun:          false,
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 10 * time.Minute,
		},
		SomeMysqlSpecialOption: true,
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) (*gorm.DB, error) {
				return database.NewMysqlDB(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
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

// GetOptions 返回所有数据库相关的选项
func (d *DatabaseOptions) GetOptions() []fx.Option {
	return d.container
}

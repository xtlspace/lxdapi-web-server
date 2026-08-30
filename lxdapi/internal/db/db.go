package db

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"lxdapi/internal/core"
	"lxdapi/models"
	"strings"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

func Init() error {
	var err error
	cfg := core.GlobalConfig.System.Database

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User,
			cfg.MySQL.Password,
			cfg.MySQL.Host,
			cfg.MySQL.Port,
			cfg.MySQL.Database,
		)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.User,
			cfg.Postgres.Password,
			cfg.Postgres.Database,
			cfg.Postgres.SSLMode,
		)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		dsn := cfg.SQLite.Path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return err
		}
		DB, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	}

	if err != nil {
		return err
	}

	// 先将历史文本列 traffic_usage 迁移为数值列，再 AutoMigrate 对齐模型，
	// 否则 GORM 对类型差异列的重建/变更可能失败或产生意外结果。
	migrateLegacyTrafficUsage()

	if err := DB.AutoMigrate(
		&models.Container{},
		&models.PortMappingV4{},
		&models.PortMappingV6{},
		&models.NATConfigV4{},
		&models.NATConfigV6{},
		&models.IPv4Binding{},
		&models.IPv4Pool{},
		&models.Task{},
		&models.ContainerCredential{},
		&models.Template{},
		&models.BrandSettings{},
		&models.PortRangeConfig{},
		&models.IPPoolSettings{},
		&models.AccessToken{},
		&models.IPv6NeighborConfig{},
	); err != nil {
		return err
	}

	dropLegacyColumns()
	return nil
}

// migrateLegacyTrafficUsage 将旧版 traffic_usage 文本列（如 "1.2345"）转为原生数值列。
// 已存在的表且列为文本类型时才迁移；新库/已数值化则直接跳过（幂等）。
func migrateLegacyTrafficUsage() {
	if !DB.Migrator().HasTable(&models.Container{}) {
		return
	}

	isText := false
	if types, err := DB.Migrator().ColumnTypes(&models.Container{}); err == nil {
		for _, t := range types {
			if t.Name() != "traffic_usage" {
				continue
			}
			dt := strings.ToUpper(t.DatabaseTypeName())
			if strings.Contains(dt, "CHAR") || strings.Contains(dt, "TEXT") || dt == "VARCHAR" {
				isText = true
			}
			break
		}
	}
	if !isText {
		return
	}

	switch DB.Dialector.Name() {
	case "mysql":
		// 旧文本值（含 ""）由 MySQL 自动转换，"" -> 0
		if err := DB.Exec("ALTER TABLE containers MODIFY COLUMN traffic_usage DOUBLE NULL").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (mysql): %v\n", err)
		}
	case "postgres":
		// 空串/NULL 无法直接 ::double precision，用 NULLIF 兜底为 NULL
		if err := DB.Exec("ALTER TABLE containers ALTER COLUMN traffic_usage TYPE DOUBLE PRECISION USING (NULLIF(traffic_usage, '')::double precision)").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (postgres): %v\n", err)
		}
	default:
		// SQLite：新建 REAL 列 -> 复制数据 -> 删旧列 -> 改名
		if err := DB.Exec("ALTER TABLE containers ADD COLUMN traffic_usage_new REAL").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (sqlite add): %v\n", err)
			return
		}
		if err := DB.Exec("UPDATE containers SET traffic_usage_new = CAST(traffic_usage AS REAL)").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (sqlite copy): %v\n", err)
			return
		}
		if err := DB.Exec("ALTER TABLE containers DROP COLUMN traffic_usage").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (sqlite drop): %v\n", err)
			return
		}
		if err := DB.Exec("ALTER TABLE containers RENAME COLUMN traffic_usage_new TO traffic_usage").Error; err != nil {
			fmt.Printf("迁移 traffic_usage 列失败 (sqlite rename): %v\n", err)
			return
		}
	}
}

func dropLegacyColumns() {
	legacy := []string{"cpu_usage", "memory_usage", "disk_usage", "traffic_usage_raw"}
	for _, col := range legacy {
		if DB.Migrator().HasColumn(&models.Container{}, col) {
			if err := DB.Migrator().DropColumn(&models.Container{}, col); err != nil {
				fmt.Printf("删除 containers.%s 旧列失败: %v\n", col, err)
			}
		}
	}
}


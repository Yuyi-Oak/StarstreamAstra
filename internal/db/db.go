package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"StarstreamAstra/internal/model"
)

type DBConn struct {
	Gorm  *gorm.DB
	sqlDB *sql.DB
}

func InitDB(dsn string, useZap bool, zapLogger *zap.Logger, loggerChoice string) (*DBConn, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	var gormLogger logger.Interface
	switch strings.ToLower(loggerChoice) {
	case "silent":
		gormLogger = logger.Default.LogMode(logger.Silent)
	case "error":
		gormLogger = logger.Default.LogMode(logger.Error)
	case "warn":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "debug":
		gormLogger = logger.Default.LogMode(logger.Info)
	default:
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	var err error
	if useZap {
		if zapLogger == nil {
			if strings.ToLower(loggerChoice) == "debug" {
				zapLogger, err = zap.NewDevelopment()
			} else {
				zapLogger, err = zap.NewProduction()
			}
			if err != nil {
				return nil, fmt.Errorf("failed to build zap logger: %w", err)
			}
		}

		zgorm := zapgorm2.New(zapLogger.Named("gorm"))
		zgorm.SetAsDefault()
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: zgorm,
		})
		if err != nil {
			return nil, fmt.Errorf("gorm open error with zap: %w", err)
		}

		sqlDB, err := gdb.DB()
		if err != nil {
			return nil, fmt.Errorf("get sql.DB error: %w", err)
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("ping db error: %w", err)
		}

		if err := gdb.AutoMigrate(
			&model.User{},
			&model.Node{},
			&model.VM{},
			&model.Order{},
		); err != nil {
			zapLogger.Sugar().Warnf("auto migrate warning: %v", err)
		}

		return &DBConn{Gorm: gdb, sqlDB: sqlDB}, nil
	}

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open error: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB error: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db error: %w", err)
	}

	if err := gdb.AutoMigrate(
		&model.User{},
		&model.Node{},
		&model.VM{},
		&model.Order{},
	); err != nil {
		log.Printf("Auto migrate warning: %v", err)
	}

	return &DBConn{Gorm: gdb, sqlDB: sqlDB}, nil
}

func (d *DBConn) Close() error {
	if d == nil || d.sqlDB == nil {
		return nil
	}
	return d.sqlDB.Close()
}

func RunMigrations(databaseURL string, migrationsPath string) error {
	if strings.TrimSpace(databaseURL) == "" {
		return errors.New("empty databaseURL for migrations")
	}
	if strings.TrimSpace(migrationsPath) == "" {
		return errors.New("empty migrationsPath")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("migration up error: %w", err)
	}
	return nil
}

// Package database opens, migrates and seeds the SQLite database.
package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	serverutils "nex-server/utils"
	"nextduel-svr/config"
	"nextduel-svr/models"
	"nextduel-svr/questions"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens the database at cfg.DBPath, migrates it and seeds defaults.
func Open(cfg config.Config) (*gorm.DB, error) {
	if dir := filepath.Dir(cfg.DBPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	// WAL lets reads proceed during writes; busy_timeout absorbs short
	// write contention instead of failing with SQLITE_BUSY.
	dsn := cfg.DBPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := initDefaultData(db, cfg); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Quest{},
		&models.Players{},
		&models.Match{},
		&models.MatchPlayer{},
		&models.MatchRound{},
		&models.RoundPlayer{},
		&models.MatchEventLog{},
		&models.ScoreEvent{},
	)
}

func initDefaultData(db *gorm.DB, cfg config.Config) error {
	if err := questions.NewBank(db).Seed(context.Background()); err != nil {
		return err
	}
	return ensureAdmin(db, cfg.AdminPassword)
}

// ensureAdmin creates the admin account only when ADMIN_PASSWORD is set, and
// warns loudly if a legacy admin/admin account is still present.
func ensureAdmin(db *gorm.DB, password string) error {
	var admin models.User
	err := db.Where("username = ?", "admin").First(&admin).Error
	switch {
	case err == nil:
		if serverutils.CheckPassword(admin.Password, "admin") {
			slog.Warn("SECURITY: the admin account still uses the default password 'admin' — change or disable it")
		}
		return nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return err
	case password == "":
		return nil
	}

	hashed, err := serverutils.HashPassword(password)
	if err != nil {
		return err
	}
	slog.Info("created admin account from ADMIN_PASSWORD")
	return db.Create(&models.User{
		Username: "admin",
		Email:    "admin@localhost",
		Password: hashed,
		Role:     "admin",
		IsActive: true,
	}).Error
}

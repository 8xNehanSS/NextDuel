package main

import (
	serverutils "nex-server/utils"
	"nextduel-svr/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("nextduel.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = migrateDatabase(db)
	if err != nil {
		return nil, err
	}
	err = initDefaultData(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Quest{},
		&models.Players{},
		&models.Match{},
		&models.MatchPlayer{},
		&models.MatchRound{},
		&models.RoundPlayer{},
		&models.MatchEventLog{},
		&models.RoundPlayer{},
		&models.ScoreEvent{},
	)
}

func initDefaultData(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hashedPassword, err := serverutils.HashPassword("admin")
		if err != nil {
			return err
		}
		admin := models.User{
			Username: "admin",
			Email:    "",
			Password: hashedPassword,
			Role:     "admin",
			IsActive: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	}
	return nil
}

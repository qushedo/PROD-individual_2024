package database

import (
	"backend-qushedo/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DSN")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // Logger: logger.Default.LogMode(logger.Info)
	if err == nil {
		DB = database
		err = database.AutoMigrate(
			&models.User{},
			&models.Travel{},
			&models.Location{},
			&models.TravelMember{},
			&models.Invite{},
			&models.Note{},
			&models.Transaction{},
		)

		log.Println("Connected to database")
	} else {
		panic(err)
	}
}

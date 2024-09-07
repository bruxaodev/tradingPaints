package dataBase

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/bruxaodev/tradingPaints/internal/schemas"

	"github.com/bruxaodev/go-logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *Database

func NewSql() (*Database, error) {
	logger := logger.New("SQLITE", logger.Yellow)
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(usr.HomeDir, "Documents", "iracing", "paint")
	dbFile := filepath.Join(dbPath, "tradingPaints.db")

	_, err = os.Stat(dbFile)
	if os.IsNotExist(err) {
		logger.Info("DATABASE NOT FOUND, creating a new one")
		//create a directory if not exists
		err = os.MkdirAll(dbPath, os.ModePerm)
		if err != nil {
			logger.Error("Error creating database directory", err)
			return nil, err
		}

		file, err := os.Create(dbFile)
		if err != nil {
			logger.Error("Error creating database file", err)
			return nil, err
		}
		file.Close()
	}
	my_db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		logger.Error("Error on connect to database", err)
	}

	err = my_db.AutoMigrate(&schemas.Paint{})
	if err != nil {
		logger.Error("Error on migrate database", err)
		return nil, err
	}
	db = &Database{Db: my_db}
	return db, nil
}

func GetDb() *Database {
	return db
}

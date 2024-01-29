package sql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB

	ip       string
	port     string
	dbName   string
	user     string
	password string
}

func New(ip string, port string, dbName string, user string, password string) *Database {
	return &Database{
		DB: nil,

		ip:       ip,
		port:     port,
		dbName:   dbName,
		user:     user,
		password: password,
	}
}
func (db *Database) Open() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.user, db.password, db.ip, db.port, db.dbName)
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	db.DB = gormDB

	return nil
}

func (d *Database) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}

	return db.Close()
}

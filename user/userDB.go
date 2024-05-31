package user

import (
	"log"
	"os"

	"github.com/objectbox/objectbox-go/objectbox"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen

// For ObjectBox (Deprecated)
type User struct {
	Id             uint64
	UID            string `objectbox:"unique"`
	Salt           string // may be empty. In Argon2 the HashedPassword contains the salt.
	HashedPassword string
	PrivateKey     string
	PublicKey      string
	CommandToUser  string
	PushUrl        string
	LocationData   []string // elements must be string-encoded JSON structures
	Pictures       []string // elements are base64 encoded encrypted images
}

type FMDDB struct {
	DB *gorm.DB
}

// For ObjectBox
type DB struct {
	Id      uint64
	Setting string `objectbox:"unique"`
	Value   string
}

// For GORM (SQL)
// User Table
type FMDUser struct {
	Id             uint64 `gorm:"primaryKey"`
	UID            string `gorm:"uniqueIndex"`
	Salt           string // may be empty. In Argon2 the HashedPassword contains the salt.
	HashedPassword string
	PrivateKey     string
	PublicKey      string
	CommandToUser  string
	PushUrl        string
	Locations      []Location `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Pictures       []Picture  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

// Location Table of the Users
type Location struct {
	Id       uint64 `gorm:"primaryKey"`
	UserID   uint64
	Position string // elements must be string-encoded JSON structures
}

// Picture Table for the Users
type Picture struct {
	Id      uint64 `gorm:"primaryKey"`
	UserID  uint64
	Content string // elements must be string-encoded JSON structures
}

// Settings Table GORM (SQL)
type DBSetting struct {
	Id      uint64 `gorm:"primaryKey"`
	Setting string `gorm:"uniqueIndex"`
	Value   string
}

func initSQLite(path string) *FMDDB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			IgnoreRecordNotFoundError: false, // Ignore ErrRecordNotFound error for logger
		},
	)
	db, _ := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: newLogger,
	})
	db.AutoMigrate(&FMDUser{}, &Location{}, &Picture{})
	db.AutoMigrate(&DBSetting{})
	return &FMDDB{DB: db}
}

func (db *FMDDB) GetLastID() int {
	var user FMDUser
	db.DB.Last(&user)
	if user.Id == 0 {
		return -1
	}
	return int(user.Id)
}

func (db *FMDDB) GetByID(id string) *FMDUser {
	var user = FMDUser{UID: id}
	db.DB.Preload("Locations").Preload("Pictures").Where(&user).First(&user)
	if user.Id == 0 {
		return nil
	}
	return &user
}

func (db *FMDDB) Save(value interface{}) {
	db.DB.Save(value)
}

func (db *FMDDB) Create(value interface{}) {
	db.DB.Create(value)
}

func (db *FMDDB) Delete(value interface{}) {
	db.DB.Delete(value)
}

// Deprecated for migration of ObjectBox DB only.
func initObjectBox(path string) *UserBox {
	ob, _ := objectbox.NewBuilder().MaxSizeInKb(10 * 1048576).Model(ObjectBoxModel()).Directory(path).Build()

	u := BoxForUser(ob)
	dbc := BoxForDB(ob)
	dbc.updateDB(u)

	return u
}

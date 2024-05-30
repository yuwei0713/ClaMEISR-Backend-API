package routine

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MeisrDB *gorm.DB
var BackStageDB *gorm.DB

func ConnectDB(DBName string, db **gorm.DB) {
	dbName := os.Getenv(DBName + "_DATABASE")
	dbUser := os.Getenv("MYSQL_USERNAME")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbPassword := os.Getenv("MYSQL_PASSWORD")

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	*db, err = gorm.Open(mysql.Open(dbURI), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
}

func InitDB() {
	ConnectDB("MEISR", &MeisrDB)
	ConnectDB("BACKSTAGE", &BackStageDB)
}

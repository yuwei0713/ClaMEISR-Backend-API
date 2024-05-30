package main

import (
	"fmt"
	Routers "ginapi/routers"
	Routines "ginapi/routine"
	"log"

	"github.com/gin-contrib/cors"
)

func main() {
	Routines.InitDB()

	var db = Routines.MeisrDB
	if db == nil {
		log.Fatal("Database is nil. Make sure it is properly initialized.")
	}
	// // Create route
	router := Routers.InitRouters()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"PUT", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"X-Requested-With", "Content-Type"},
	}))

	// Run server
	errorMessage := router.Run(":9000")
	if errorMessage != nil {
		fmt.Println("Service Error")
	}
}

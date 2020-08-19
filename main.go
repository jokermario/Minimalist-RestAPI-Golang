package main

import (
	"github.com/Minimalist-RestAPI-Golang/config"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func init() {
	//loads value from .env into the system
	if err := godotenv.Load("application.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main(){
	logger := log.New(os.Stdout, "REST-API -- ", log.LstdFlags | log.Lshortfile)
	conf := config.NewConfig()

	a := App{}
	a.Initialize(conf.DbUsername, conf.DbPassword, conf.DbName)

	logger.Println("Server is starting")
	a.Run(conf.ServerPort)
	logger.Println("Server started")
}

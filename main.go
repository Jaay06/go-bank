package main

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv () {
	err :=godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func main(){

	LoadEnv()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil{
		log.Fatal(err)
	}

	// test if server is active
	// fmt.Printf("%+v\n", store)

	server := NewAPIServer(":3000", store)
	server.Run()
}
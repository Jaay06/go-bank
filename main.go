package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv () {
	err :=godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func seedAccount (store Storage, fname, lname, pw string) *Account{
	acc, err := NewAccount(fname, lname, pw)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new accoun =>", acc.Number)

	return acc
}


func seedAccounts(s Storage){
	seedAccount(s, "Jaay", "lastname", "password")
}

func main(){

	LoadEnv()

	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil{
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}


	server := NewAPIServer(":3000", store)
	server.Run()
}
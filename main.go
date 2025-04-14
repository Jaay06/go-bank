package main

import (
	"log"
)

func main(){
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	// test if server is active
	//fmt.Printf("%+v\n", store)

	server := NewAPIServer(":3000", store)
	server.Run()
}
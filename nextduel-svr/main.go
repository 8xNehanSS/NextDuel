package main

import (
	"log"
	"nextduel-svr/routes"
	"nextduel-svr/server"
)

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	s := server.NewServer()

	s.SetDatabase(db)
	routes.RegisterRoutes(s)

	log.Fatal(s.Start())
}
